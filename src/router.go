package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nsw42/piaf/mediadir"
)

type ViewStyle string

const (
	ViewByAge     ViewStyle = "by-age"
	ViewInFolders ViewStyle = "by-folders"
)

type GetStatusResponse struct {
	Status        string   `json:"state"`
	NowPlaying    *string  `json:"now_playing"`
	TrackDuration *int     `json:"duration"`
	Position      *float64 `json:"position"`
	Speed         string   `json:"speed"`
	Volume        int      `json:"volume"` // 0 <= Volume <= 100
}

type TemplatePageArgs struct {
	RequestPath              string
	RequestPathElts          [][2]string
	EnableRemoteControl      bool
	EnableBrowserPlayback    bool
	EnableSpeedControl       bool
	IncludeFooterPauseResume bool
	ViewStyle                string
}

var phoneAddressHistoryFilePath string
var phoneAddressHistory = make([]string, 0)
var podcastSortOrder []directoryAge = nil
var podcastViewOffset = 0
var oldestFile *mediadir.MediaFile // Remember the last file that was top of the by-age view

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Unable to find user home directory: cannot load history of upload addresses")
		return
	}
	phoneAddressHistoryFilePath = fmt.Sprintf("%s/.piafrc", home)
	historyFile, err := os.Open(phoneAddressHistoryFilePath)
	if err != nil {
		return
	}
	defer historyFile.Close()

	contents, _ := io.ReadAll(historyFile)
	json.Unmarshal(contents, &phoneAddressHistory)
}

func ConfigureRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", rootHandler)
	router.POST("/viewstyle", viewstyleHandler)
	router.GET("/media/*path", indexPageHandler)
	router.Static("/mediafile", Args.MediaParentDirectory+"/Unplayed")
	router.DELETE("/mediafile/*path", markPlayedHandler)
	router.GET("/player/control", controlPageHandler)
	router.PUT("/player/play/*path", playHandler)
	router.PUT("/player/pause", pauseHandler)
	router.PUT("/player/resume", resumeHandler)
	router.PUT("/player/seek", seekHandler)
	router.GET("/player/status", getPlayerStatusHandler)
	router.PUT("/player/speed", speedHandler)
	router.PUT("/player/volume", volumeHandler)
	router.POST("/phone", sendToPhoneHandler)
	configureAssetsForRouter(router, "/assets")
	return router
}

func splitPath(path string) []string {
	pathElts := make([]string, 0)
	for elt := range strings.SplitSeq(path, "/") {
		if elt != "" {
			pathElts = append(pathElts, elt)
		}
	}

	return pathElts
}

func getUriPathElements(c *gin.Context) (string, []string) {
	// basically splits on /, but removes empty elements, to ensure that
	// http://server/path//subdir doesn't cause headaches
	path := c.Param("path")
	path, err := url.PathUnescape(path)
	if err != nil {
		return "", []string{}
	}

	return path, splitPath(path)
}

func findMediaDir(pathElts []string) *mediadir.MediaDirectory {
	// pathElts must only consist of the directories:
	// any trailing file must have been removed by the caller
	search := Media.Contents
	for _, elt := range pathElts {
		var ok bool
		search, ok = search.SubDirectories[elt]
		if !ok {
			return nil
		}
	}
	return search
}

func findMediaFile(pathElts []string) *mediadir.MediaFile {
	mediaDir := findMediaDir(pathElts[:len(pathElts)-1])
	if mediaDir == nil {
		return nil
	}
	return mediaDir.Files[pathElts[len(pathElts)-1]]
}

func formatPathElts(pathElts []string) [][2]string {
	linkPathElts := make([][2]string, 1+len(pathElts)) // [0] = link dest (or "" if none), [1] = text
	linkPathElts[0][1] = "Root"
	if len(pathElts) == 0 {
		linkPathElts[0][0] = "" // no link
	} else {
		linkDest := "/media"
		linkPathElts[0][0] = linkDest
		for i, pathElt := range pathElts {
			if i == len(pathElts)-1 {
				linkPathElts[i+1][0] = ""
			} else {
				linkDest = linkDest + "/" + pathElt
				linkPathElts[i+1][0] = linkDest
			}
			linkPathElts[i+1][1] = pathElt
		}
	}

	return linkPathElts
}

func rootHandler(c *gin.Context) {
	c.Header("Cache-Control", "max-age=60, must-revalidate")
	c.Redirect(http.StatusMovedPermanently, "/media/")
}

type directoryAge struct {
	dir   *mediadir.MediaDirectory
	files []*mediadir.MediaFile
	age   time.Time
}

func getEpisodeTime(mf *mediadir.MediaFile) time.Time {
	// Not mf.ModTime, because that's the file modification time
	// We want to know when the podcast episode was published, which is (probably) encoded in its filename
	slash := strings.LastIndex(mf.RelativePath, "/")
	leaf := mf.RelativePath[slash+1:]
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	if re.MatchString(leaf) {
		// leaf starts with YYYY-MM-DD as expected
		yyyy, _ := strconv.Atoi(leaf[0:4])
		mm, _ := strconv.Atoi(leaf[5:7])
		dd, _ := strconv.Atoi(leaf[8:10])
		rtn := time.Date(yyyy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
		return rtn
	} else {
		log.Println(leaf, "does not match YYYY-MM-DD")
		return mf.ModTime // it's the best info we've got
	}
}

func getPodcastDirectoryAge(dir *mediadir.MediaDirectory) directoryAge {
	oldest := time.Now()
	files := make([]*mediadir.MediaFile, 0, len(dir.Files))
	for _, mf := range dir.Files {
		files = append(files, mf)
	}
	slices.SortFunc(files, func(mf1, mf2 *mediadir.MediaFile) int {
		return strings.Compare(mf1.RelativePath, mf2.RelativePath)
	})

	for _, mf := range dir.Files {
		fileTime := getEpisodeTime(mf)
		if fileTime.Before(oldest) {
			oldest = fileTime
		}
	}

	return directoryAge{dir, files, oldest}
}

func sortPodcastSeriesByAge() []directoryAge {
	podcastAges := make([]directoryAge, 0)
	dirsToCheck := make([]*mediadir.MediaDirectory, 1)
	dirsToCheck[0] = Media.Contents
	for len(dirsToCheck) > 0 {
		var dir *mediadir.MediaDirectory
		dir, dirsToCheck = dirsToCheck[0], dirsToCheck[1:]
		age := getPodcastDirectoryAge(dir)
		podcastAges = append(podcastAges, age)
		for d := range maps.Values(dir.SubDirectories) {
			dirsToCheck = append(dirsToCheck, d)
		}
	}
	slices.SortFunc(podcastAges, func(dir1age, dir2age directoryAge) int {
		return dir1age.age.Compare(dir2age.age)
	})
	return podcastAges
}

func indexPageHandler(c *gin.Context) {
	path, pathElts := getUriPathElements(c)

	viewStyle, err := c.Cookie("piaf-view-style")
	if (err != nil) || (viewStyle != string(ViewByAge) && viewStyle != string(ViewInFolders)) {
		viewStyle = string(ViewInFolders)
	}

	mediaDir := findMediaDir(pathElts)
	// traverse our media tree looking for the requested directory
	if mediaDir == nil {
		// Redirect to root if the directory does not exist
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Redirect(http.StatusTemporaryRedirect, "/media/")
		return
	}

	mediaDir.RefreshAndGetMetadata() // This ensures our cache is up-to-date
	if mediaDir.RelativePath != "." && mediaDir.TotalDurationSeconds == 0 {
		// Redirect to root if a sub-directory is empty
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Redirect(http.StatusTemporaryRedirect, "/media/")
		return
	}

	var filesInViewOrder []*mediadir.MediaFile
	if path == "/" && viewStyle == string(ViewByAge) {
		// Construct a fake media directory containing all files
		var podcasts []directoryAge
		if podcastSortOrder == nil {
			podcasts = sortPodcastSeriesByAge()
			if Media.Contents.TotalDurationSeconds > 0 {
				// We've finally finished building the index, so we can save the sort order
				podcastSortOrder = podcasts
			}
		} else {
			// we have an established sort order - use it
			podcasts = podcastSortOrder
			if oldestFile != nil {
				currentViewFiles := slices.Collect(maps.Values(podcasts[podcastViewOffset].dir.Files))
				if !slices.Contains(currentViewFiles, oldestFile) {
					// The file has gone away - so move to the next column
					podcastViewOffset += 1
					oldestFile = nil
				}
			}
		}
		if len(podcasts) > 0 {
			// Only do anything if there are episodes found
			podcastIndexes := make([]int, len(podcasts))
			done := false
			filesInViewOrder = make([]*mediadir.MediaFile, 0)
			for !done {
				done = true // until we decide otherwise
				// make one pass over the podcasts, offset by the current view offset
				for i := range len(podcasts) {
					i = (podcastViewOffset + i) % len(podcasts)
					podcast := podcasts[i]
					fileIndex := podcastIndexes[i]
					if fileIndex < len(podcast.files) {
						filesInViewOrder = append(filesInViewOrder, podcast.files[fileIndex])
						podcastIndexes[i] += 1
						done = false
					}
				}
			}
			oldestFile = filesInViewOrder[0]
		}
	} else {
		// files are just the ones from this directory
		filesInViewOrder = slices.Collect(maps.Values(mediaDir.Files))
		// Sort??
	}

	pageTemplate, err := getTemplate("index.templ")
	if err != nil || pageTemplate == nil {
		log.Println("Unable to read template index.templ", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	linkPathElts := formatPathElts(pathElts)

	pageArgs := struct {
		TemplatePageArgs
		SubDirectories      map[string]*mediadir.MediaDirectory
		Files               []*mediadir.MediaFile
		TotalDurationString string
		PhoneAddressHistory []string
	}{
		TemplatePageArgs: TemplatePageArgs{
			RequestPath:              path,
			RequestPathElts:          linkPathElts,
			EnableRemoteControl:      Args.EnableRemoteControl,
			EnableBrowserPlayback:    Args.EnableBrowserPlayback,
			EnableSpeedControl:       Args.EnableSpeedControl,
			IncludeFooterPauseResume: true,
			ViewStyle:                viewStyle,
		},
		SubDirectories:      mediaDir.SubDirectories,
		Files:               filesInViewOrder,
		TotalDurationString: mediaDir.TotalDurationString,
		PhoneAddressHistory: phoneAddressHistory,
	}
	err = pageTemplate.Execute(c.Writer, pageArgs)
	if err != nil {
		log.Println("Failed executing template:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func viewstyleHandler(c *gin.Context) {
	newStyle := c.PostForm("style")
	if (newStyle != string(ViewByAge)) && (newStyle != string(ViewInFolders)) {
		c.Status(http.StatusBadRequest)
		return
	}
	c.SetCookie("piaf-view-style", newStyle, 60*60*24*365*10, "/", "", false, false)
	c.Redirect(http.StatusMovedPermanently, "/")
}

func controlPageHandler(c *gin.Context) {
	type ControlPageArgs struct {
		TemplatePageArgs
		CurrentStatus string
	}
	var pageArgs ControlPageArgs
	if MediaPlayer.NowPlaying == nil {
		pageArgs = ControlPageArgs{
			TemplatePageArgs: TemplatePageArgs{
				RequestPath:              "",
				RequestPathElts:          make([][2]string, 0),
				EnableRemoteControl:      Args.EnableRemoteControl,
				EnableBrowserPlayback:    false,
				EnableSpeedControl:       Args.EnableSpeedControl,
				IncludeFooterPauseResume: false,
			},
			CurrentStatus: MediaPlayer.State.String(),
		}
	} else {
		pageArgs = ControlPageArgs{
			TemplatePageArgs: TemplatePageArgs{
				RequestPath:           MediaPlayer.NowPlaying.RelativePath,
				RequestPathElts:       formatPathElts(splitPath(MediaPlayer.NowPlaying.RelativePath)),
				EnableRemoteControl:   Args.EnableRemoteControl,
				EnableBrowserPlayback: false,
				EnableSpeedControl:    Args.EnableSpeedControl,
			},
			CurrentStatus: MediaPlayer.State.String(),
		}
	}
	pageTemplate, err := getTemplate("control.templ")
	if err != nil || pageTemplate == nil {
		log.Println("Unable to read template control.templ", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	err = pageTemplate.Execute(c.Writer, pageArgs)
	if err != nil {
		log.Println("Failed executing template:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func getPlayerStatusHandler(c *gin.Context) {
	var nowPlaying *string = nil
	var duration *int = nil
	var posValue float64
	var position *float64 = nil
	if MediaPlayer.NowPlaying != nil {
		nowPlaying = &MediaPlayer.NowPlaying.RelativePath
		duration = &MediaPlayer.NowPlaying.DurationSeconds
		posValue = MediaPlayer.GetPosition().Seconds()
		position = &posValue
	}
	response := GetStatusResponse{
		Status:        MediaPlayer.State.String(),
		NowPlaying:    nowPlaying,
		TrackDuration: duration,
		Position:      position,
		Speed:         MediaPlayer.SpeedString,
		Volume:        MediaPlayer.Volume,
	}
	c.JSON(http.StatusOK, response)
}

func markPlayedHandler(c *gin.Context) {
	_, pathElts := getUriPathElements(c)
	if len(pathElts) == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	file := findMediaFile(pathElts)
	if file == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if err := Media.MarkFilePlayed(file); err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func playHandler(c *gin.Context) {
	_, pathElts := getUriPathElements(c) // TODO: This would make more sense as a query param than a uri param
	if len(pathElts) == 0 {
		// No file to play
		c.Status(http.StatusNotFound)
		return
	}

	file := findMediaFile(pathElts)
	if file == nil {
		c.Status(http.StatusNotFound)
		return
	}

	MediaPlayer.Play(file, Args.EnableSpeedControl, func() {
		Media.MarkFilePlayed(file)
	})

	c.Status(http.StatusNoContent)
}

func pauseHandler(c *gin.Context) {
	if MediaPlayer.State == PlayerStatePlaying {
		MediaPlayer.Pause()
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusConflict)
	}
}

func resumeHandler(c *gin.Context) {
	if MediaPlayer.State == PlayerStatePaused {
		MediaPlayer.Resume()
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusConflict)
	}
}

func seekHandler(c *gin.Context) {
	arg := c.Query("p")
	positionSeconds, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		log.Println("Unable to parse ", arg)
		c.Status(http.StatusBadRequest)
		return
	}
	if err = MediaPlayer.SetPosition(time.Duration(positionSeconds * float64(time.Second))); err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func speedHandler(c *gin.Context) {
	if !Args.EnableSpeedControl {
		c.Status(http.StatusConflict)
		return
	}
	speedStr := c.Query("v")
	err := MediaPlayer.SetSpeed(speedStr)
	if err != nil {
		rtn := make(map[string]string, 1)
		rtn["error"] = err.Error()
		c.JSON(http.StatusBadRequest, rtn)
		return
	}
	c.Status(http.StatusNoContent)
}

func volumeHandler(c *gin.Context) {
	volStr := c.Query("v")
	vol, err := strconv.Atoi(volStr)
	if err != nil || vol < 0 || vol > 100 {
		c.Status(http.StatusBadRequest)
		return
	}
	MediaPlayer.SetVolume(vol)
	c.Status(http.StatusNoContent)
}

func newFileUploadRequest(url string, file *mediadir.MediaFile, filename string) (*http.Request, error) {
	// Read the file to send
	handle, err := os.Open(file.Path)
	if err != nil {
		return nil, err
	}
	contents, err := io.ReadAll(handle)
	if err != nil {
		return nil, err
	}
	handle.Close()

	// Prepare the request body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	quotedFilename := strings.ReplaceAll(filename, "\"", "%22")
	part, err := writer.CreateFormFile("files[]", quotedFilename)
	if err != nil {
		return nil, err
	}
	part.Write(contents)
	writer.WriteField("path", "/")
	writer.Close()

	// Construct the request
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}

func sendToPhoneHandler(c *gin.Context) {
	var requestParams struct {
		File         string `json:"file"`
		PhoneAddress string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	pathElts := splitPath(requestParams.File)
	file := findMediaFile(pathElts)
	if file == nil {
		c.Status(http.StatusNotFound)
		return
	}

	updatePhoneAddressHistory(requestParams.PhoneAddress)

	url := fmt.Sprintf("http://%s/upload", requestParams.PhoneAddress)
	request, err := newFileUploadRequest(url, file, pathElts[len(pathElts)-1])
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		c.Status(http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	c.Status(http.StatusNoContent)
}

func updatePhoneAddressHistory(phoneAddress string) {
	var newHistory = []string{phoneAddress}
	for _, address := range phoneAddressHistory {
		if !slices.Contains(newHistory, address) {
			newHistory = append(newHistory, address)
			if len(newHistory) == 5 {
				break
			}
		}
	}
	phoneAddressHistory = newHistory

	if phoneAddressHistoryFilePath != "" {
		marshalled, err := json.MarshalIndent(phoneAddressHistory, "", "  ")
		if err == nil {
			os.WriteFile(phoneAddressHistoryFilePath, marshalled, 0644)
		}
	}
}

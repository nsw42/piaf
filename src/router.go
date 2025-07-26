package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nsw42/piaf/mediadir"
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
	EnableSpeedControl       bool
	IncludeFooterPauseResume bool
}

func ConfigureRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/media/") })
	router.GET("/media/*path", getIndexPageHandler)
	router.GET("/player/control", getControlPageHandler)
	router.PUT("/player/play/*path", playHandler)
	router.PUT("/player/pause", pauseHandler)
	router.PUT("/player/resume", resumeHandler)
	router.PUT("/player/seek", seekHandler)
	router.GET("/player/status", getPlayerStatusHandler)
	router.PUT("/player/speed", speedHandler)
	router.PUT("/player/volume", volumeHandler)
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
	path, err := url.QueryUnescape(path)
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

func getIndexPageHandler(c *gin.Context) {
	path, pathElts := getUriPathElements(c)

	mediaDir := findMediaDir(pathElts)
	// traverse our media tree looking for the requested directory
	if mediaDir == nil {
		c.String(http.StatusNotFound, path+" not found")
		return
	}

	// TODO: Move template loading back into ConfigureRouter()
	pageTemplate, err := getTemplate("index.templ")
	if err != nil || pageTemplate == nil {
		log.Println("Unable to read template index.templ", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	linkPathElts := formatPathElts(pathElts)

	pageArgs := struct {
		TemplatePageArgs
		MediaDir *mediadir.MediaDirectory
	}{
		TemplatePageArgs: TemplatePageArgs{
			RequestPath:              path,
			RequestPathElts:          linkPathElts,
			EnableSpeedControl:       Args.EnableSpeedControl,
			IncludeFooterPauseResume: true,
		},
		MediaDir: mediaDir,
	}
	err = pageTemplate.Execute(c.Writer, pageArgs)
	if err != nil {
		log.Println("Failed executing template:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func getControlPageHandler(c *gin.Context) {
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
				EnableSpeedControl:       Args.EnableSpeedControl,
				IncludeFooterPauseResume: false,
			},
			CurrentStatus: MediaPlayer.State.String(),
		}
	} else {
		pageArgs = ControlPageArgs{
			TemplatePageArgs: TemplatePageArgs{
				RequestPath:     MediaPlayer.NowPlaying.RelativePath,
				RequestPathElts: formatPathElts(splitPath(MediaPlayer.NowPlaying.RelativePath)),
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

func playHandler(c *gin.Context) {
	_, pathElts := getUriPathElements(c) // TODO: This would make more sense as a query param than a uri param
	if len(pathElts) == 0 {
		// No file to play
		c.Status(http.StatusNotFound)
		return
	}

	mediaDir := findMediaDir(pathElts[:len(pathElts)-1])
	file := mediaDir.Files[pathElts[len(pathElts)-1]]
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

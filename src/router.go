package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GetStatusResponse struct {
	Status     string  `json:"state"`
	NowPlaying *string `json:"now_playing"`
	Speed      string  `json:"speed"`
	Volume     int     `json:"volume"` // 0 <= Volume <= 100
}

func ConfigureRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/media/") })
	router.GET("/media/*path", getPageHandler)
	router.PUT("/player/play/*path", playHandler)
	router.PUT("/player/pause", pauseHandler)
	router.PUT("/player/resume", resumeHandler)
	router.GET("/player/status", getPlayerStatusHandler)
	router.PUT("/player/speed", speedHandler)
	router.PUT("/player/volume", volumeHandler)
	router.Static("/assets", "./assets")
	return router
}

func getUriPathElements(c *gin.Context) (string, []string) {
	// basically splits on /, but removes empty elements, to ensure that
	// http://server/path//subdir doesn't cause headaches
	pathElts := make([]string, 0)
	path := c.Param("path")
	path, err := url.QueryUnescape(path)
	if err != nil {
		return "", pathElts
	}

	for elt := range strings.SplitSeq(path, "/") {
		if elt != "" {
			pathElts = append(pathElts, elt)
		}
	}

	return path, pathElts
}

func findMediaDir(pathElts []string) *MediaDirectory {
	// pathElts must only consist of the directories:
	// any trailing file must have been removed by the caller
	search := Media
	for _, elt := range pathElts {
		var ok bool
		search, ok = search.SubDirectories[elt]
		if !ok {
			return nil
		}
	}
	return search
}

func getPageHandler(c *gin.Context) {
	path, pathElts := getUriPathElements(c)

	mediaDir := findMediaDir(pathElts)
	// traverse our media tree looking for the requested directory
	if mediaDir == nil {
		c.String(http.StatusNotFound, path+" not found")
		return
	}

	// TODO: Move template loading back into ConfigureRouter()
	PageTemplate, err := template.ParseFiles("index.templ")
	if err != nil || PageTemplate == nil {
		log.Println("Unable to read template index.templ")
		c.Status(http.StatusInternalServerError)
		return
	}

	var linkPathElts [][2]string // [0] = link dest (or "" if none), [1] = text
	linkPathElts = make([][2]string, 1+len(pathElts))
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

	pageArgs := struct {
		RequestPath     string
		RequestPathElts [][2]string
		MediaDir        *MediaDirectory
	}{
		RequestPath:     path,
		RequestPathElts: linkPathElts,
		MediaDir:        mediaDir,
	}
	err = PageTemplate.Execute(c.Writer, pageArgs)
	if err != nil {
		log.Println("Failed executing template:", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func playerStateString(state PlayerState) string {
	switch state {
	case PlayerStateStopped:
		return "stopped"
	case PlayerStatePlaying:
		return "playing"
	case PlayerStatePaused:
		return "paused"
	default:
		return "unknown"
	}
}

func getPlayerStatusHandler(c *gin.Context) {
	var nowPlaying *string = nil
	if MediaPlayer.NowPlaying != "" {
		nowPlaying = &MediaPlayer.NowPlaying
	}
	response := GetStatusResponse{
		Status:     playerStateString(MediaPlayer.State),
		NowPlaying: nowPlaying,
		Speed:      MediaPlayer.SpeedString,
		Volume:     MediaPlayer.Volume,
	}
	c.JSON(http.StatusOK, response)
}

func playHandler(c *gin.Context) {
	displayPath, pathElts := getUriPathElements(c) // TODO: This would make more sense as a query param than a uri param
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

	MediaPlayer.Play(file.Path)
	MediaPlayer.NowPlaying = displayPath

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

func speedHandler(c *gin.Context) {
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

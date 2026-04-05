package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/nsw42/piaf/mediadir"
)

type CommandLineArguments struct {
	MediaParentDirectory  string
	EnableRemoteControl   bool
	EnableBrowserPlayback bool
	EnableSpeedControl    bool
	ListenPort            int
}

var Args *CommandLineArguments
var Media *mediadir.RootMediaDirectory
var PageTemplate *template.Template
var MediaPlayer *Player

func parseArgs() *CommandLineArguments {
	mediaDir := flag.String("d", "", "play files from DIR")
	enableSpeed := flag.Bool("s", false, "enable speed controls")
	playerTypes := flag.String("c", "both", "select whether to enable 'browser', 'remote', or 'both' for playback")
	listenPort := flag.Int("p", 80, "port to listen on")
	flag.Parse()

	if *mediaDir == "" {
		// Are there Unplayed and Played directories in the current directory?
		if mediadir.IsDir("./Unplayed") && mediadir.IsDir("./Played") {
			*mediaDir = "."
		} else {
			fmt.Println("The -d command-line parameter is mandatory")
			os.Exit(1)
		}
	}

	var enableRemoteControl, enableBrowserPlayback bool
	switch strings.ToLower(*playerTypes) {
	case "browser":
		enableBrowserPlayback = true
		enableRemoteControl = false
	case "remote":
		enableBrowserPlayback = false
		enableRemoteControl = true
	case "both":
		enableBrowserPlayback = true
		enableRemoteControl = true
	default:
		fmt.Println("Valid values for -c player selection are 'browser', 'remote', or 'both' (without the quotes)")
		os.Exit(1)
	}

	args := CommandLineArguments{
		MediaParentDirectory:  *mediaDir,
		EnableRemoteControl:   enableRemoteControl,
		EnableBrowserPlayback: enableBrowserPlayback,
		EnableSpeedControl:    *enableSpeed,
		ListenPort:            *listenPort,
	}

	return &args
}

func main() {
	var err error

	Args = parseArgs()
	Media, err = mediadir.ReadMediaDir(Args.MediaParentDirectory)
	if err != nil {
		panic(err)
	}

	MediaPlayer = NewPlayer()

	router := ConfigureRouter()
	router.Run(fmt.Sprintf(":%d", Args.ListenPort))

	MediaPlayer.Close()
}

package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"

	"github.com/nsw42/piaf/mediadir"
)

type CommandLineArguments struct {
	MediaParentDirectory string
	EnableSpeedControl   bool
}

var Args *CommandLineArguments
var Media *mediadir.RootMediaDirectory
var PageTemplate *template.Template
var MediaPlayer *Player

func parseArgs() *CommandLineArguments {
	mediaDir := flag.String("d", "", "play files from DIR")
	enableSpeed := flag.Bool("s", false, "enable speed controls")
	flag.Parse()

	if *mediaDir == "" {
		fmt.Println("The -d command-line parameter is mandatory")
		os.Exit(1)
	}

	args := CommandLineArguments{
		MediaParentDirectory: *mediaDir,
		EnableSpeedControl:   *enableSpeed,
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
	router.Run(":80")

	MediaPlayer.Close()
}

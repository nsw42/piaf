package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandLineArguments struct {
	MediaParentDirectory string
}

type MediaFile struct {
	DisplayName string // the leaf, with file extension removed
	Path        string // the full path (relative to the roto media parent directory)
	Duration    string // extracted from ffmpeg output
}

type MediaDirectory struct {
	Leaf           string                     // just the final element of the path
	Path           string                     // the full path
	RelativePath   string                     // the full path relative to the root media parent directory
	SubDirectories map[string]*MediaDirectory // indexed by leaf
	Files          map[string]*MediaFile      // Each entry is a full path
}

var Media *MediaDirectory
var PageTemplate *template.Template
var MediaPlayer *Player

func parseArgs() CommandLineArguments {
	mediaDir := flag.String("d", "", "play files from DIR")
	flag.Parse()

	if *mediaDir == "" {
		fmt.Println("The -d command-line parameter is mandatory")
		os.Exit(1)
	}

	_, err := os.Lstat(*mediaDir) // Quick existence test
	if err != nil {
		fmt.Println("Unable to read music directory", *mediaDir)
		os.Exit(1)
	}

	args := CommandLineArguments{
		MediaParentDirectory: *mediaDir,
	}

	return args
}

func readMediaDir(root, parent string) *MediaDirectory {
	files, err := os.ReadDir(parent)
	if err != nil {
		log.Fatal(err)
	}

	relativePath, err := filepath.Rel(root, parent)
	if err != nil {
		log.Fatal(err)
	}

	rtn := &MediaDirectory{
		Leaf:           filepath.Base(parent),
		Path:           parent, // will be initialised to a full path
		RelativePath:   relativePath,
		SubDirectories: make(map[string]*MediaDirectory, 0),
		Files:          make(map[string]*MediaFile, 0), // will be full paths (relative to the original command-line media argument)
	}

	for _, file := range files {
		fileName := file.Name()
		subPath := fmt.Sprintf("%s/%s", parent, fileName)
		if file.IsDir() {
			subdir := readMediaDir(root, subPath)
			if subdir != nil {
				rtn.SubDirectories[file.Name()] = subdir
			}
		} else {
			ext := filepath.Ext(fileName)
			switch ext {
			case ".mp3": // , ".m4a"
				displayName := strings.TrimSuffix(fileName, ext)
				rtn.Files[fileName] = &MediaFile{
					DisplayName: displayName,
					Path:        subPath,
				}
			}
		}
	}

	if len(rtn.SubDirectories) == 0 && len(rtn.Files) == 0 {
		return nil // don't bother showing empty directories
	}

	return rtn
}

func getOneMediaInfo(file *MediaFile) {
	cmd := exec.Command("ffmpeg", "-i", file.Path)
	// the info we want is in ffmpeg stderr output
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Println("Unable to get ffmpeg stderr pipe")
		return
	}
	cmd.Start()
	// Ignore error return, because ffmpeg always exits non-zero

	bytes, err := io.ReadAll(stderrPipe)
	if err != nil {
		log.Println("Error reading ffmpeg stderr stream")
		return
	}
	output := string(bytes)
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimLeft(line, " ")
		lineWords := strings.Split(line, " ")
		if len(lineWords) > 1 && lineWords[0] == "Duration:" {
			file.Duration = strings.TrimSuffix(lineWords[1], ",")
		} else if len(lineWords) > 2 && lineWords[0] == "title" {
			title := strings.TrimLeft(line[6:], " ")
			title = strings.TrimLeft(title[2:], " ")
			file.DisplayName = title
		}
	}
	cmd.Wait()
}

func getMediaLengths(directory *MediaDirectory) {
	// handle the files in this directory
	// (so we populate the files in / first)
	for _, file := range directory.Files {
		getOneMediaInfo(file)
	}

	// now recurse for subdirectories
	for _, subdir := range directory.SubDirectories {
		getMediaLengths(subdir)
	}
}

func main() {
	args := parseArgs()
	Media = readMediaDir(args.MediaParentDirectory, args.MediaParentDirectory)
	go getMediaLengths(Media)

	MediaPlayer = NewPlayer()

	router := ConfigureRouter()
	router.Run(":80")

	MediaPlayer.Close()
}

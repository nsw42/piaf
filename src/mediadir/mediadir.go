package mediadir

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type RootMediaDirectory struct {
	RootMediaDirectory     string
	UnplayedMediaDirectory string
	PlayedMediaDirectory   string
	Contents               *MediaDirectory
}

type MediaDirectory struct {
	Leaf           string                     // just the final element of the path
	Path           string                     // the full path
	RelativePath   string                     // the full path relative to the root media parent directory
	SubDirectories map[string]*MediaDirectory // indexed by leaf
	Files          map[string]*MediaFile      // Each entry is a full path
}

type MediaFile struct {
	DisplayName     string // the leaf, with file extension removed
	Path            string // the full path
	RelativePath    string // the full path relative to the root media parent directory
	DurationString  string // extracted from ffmpeg output
	DurationSeconds int
}

func isDir(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// hh:mm:ss.ms -> int(seconds)
func parseDurationString(s string) int {
	var hh, mm, ss, i int
	var err error

	fields := strings.Split(s, ":")
	if len(fields) != 3 {
		goto Error
	}

	// hh
	hh, err = strconv.Atoi(fields[0])
	if err != nil {
		goto Error
	}

	// mm
	mm, err = strconv.Atoi(fields[1])
	if err != nil {
		goto Error
	}

	// ss
	i = strings.Index(fields[2], ".")
	if i == -1 {
		err = fmt.Errorf("cannot find . in %s", fields[2])
		goto Error
	}
	ss, err = strconv.Atoi(fields[2][:i])
	if err != nil {
		goto Error
	}

	return ((hh*60)+mm)*60 + ss

Error:
	log.Println("Cannot parse duration string", s)
	if err != nil {
		log.Println(err)
	}
	return 0

}

func ReadMediaDir(root string) (*RootMediaDirectory, error) {
	media := &RootMediaDirectory{
		RootMediaDirectory:     root,
		UnplayedMediaDirectory: filepath.Join(root, "Unplayed"),
		PlayedMediaDirectory:   filepath.Join(root, "Played"),
	}

	if !isDir(media.RootMediaDirectory) {
		return nil, fmt.Errorf("root media directory '%s' does not exist", root)
	}
	if !isDir(media.UnplayedMediaDirectory) {
		return nil, fmt.Errorf("unplayed directory '%s' does not exist", root)
	}
	if !isDir(media.PlayedMediaDirectory) {
		return nil, fmt.Errorf("played directory '%s' does not exist", root)
	}

	media.Contents = readMediaDir(media.UnplayedMediaDirectory, media.UnplayedMediaDirectory)
	go getMediaLengths(media.Contents)

	return media, nil
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
		Files:          make(map[string]*MediaFile, 0),
	}

	readMediaDirPopulate(rtn, root, parent, files)

	if len(rtn.SubDirectories) == 0 && len(rtn.Files) == 0 {
		return nil // don't bother showing empty directories
	}

	return rtn
}

func readMediaDirPopulate(mediaDir *MediaDirectory, root, parent string, files []os.DirEntry) {
	for _, file := range files {
		fileName := file.Name()
		subPath := fmt.Sprintf("%s/%s", parent, fileName)
		if file.IsDir() {
			subdir := readMediaDir(root, subPath)
			if subdir != nil {
				mediaDir.SubDirectories[file.Name()] = subdir
			}
		} else {
			ext := filepath.Ext(fileName)
			if ext == ".mp3" || ext == ".m4a" {
				displayName := strings.TrimSuffix(fileName, ext)
				relativePath := filepath.Join(mediaDir.RelativePath, fileName)
				mediaDir.Files[fileName] = &MediaFile{
					DisplayName:  displayName,
					Path:         subPath,
					RelativePath: relativePath,
				}
			}
		}
	}
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
			file.DurationString = strings.TrimSuffix(lineWords[1], ",")
			file.DurationSeconds = parseDurationString(file.DurationString)
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

func (media *RootMediaDirectory) MarkFilePlayed(file *MediaFile) {
	dest := filepath.Join(media.PlayedMediaDirectory, file.RelativePath)
	os.MkdirAll(filepath.Dir(dest), 0755)
	os.Rename(file.Path, dest)

	// Remove the file from our records
	pathElts := make([]string, 0)
	filePath := file.RelativePath
	for filePath != "" {
		dir, leaf := filepath.Split(filePath)
		pathElts = append(pathElts, leaf)
		filePath = strings.TrimSuffix(dir, string(filepath.Separator))
	}
	if len(pathElts) == 0 {
		panic("error splitting path " + file.RelativePath)
	}

	mediaDir := media.Contents
	for _, dir := range slices.Backward(pathElts[1:]) {
		mediaDir = mediaDir.SubDirectories[dir]
	}
	delete(mediaDir.Files, pathElts[0])
}

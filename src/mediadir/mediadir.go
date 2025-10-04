package mediadir

import (
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type RootMediaDirectory struct {
	RootMediaDirectory     string
	UnplayedMediaDirectory string
	PlayedMediaDirectory   string
	Contents               *MediaDirectory
}

type MediaDirectory struct {
	Root                 string
	Path                 string                     // the full path
	Leaf                 string                     // just the final element of the path
	RelativePath         string                     // the full path relative to the root media parent directory
	SubDirectories       map[string]*MediaDirectory // indexed by leaf
	Files                map[string]*MediaFile      // Each entry is a full path
	TotalDurationString  string
	TotalDurationSeconds int
	ModTime              time.Time
}

type MediaFile struct {
	DisplayName     string // the leaf, with file extension removed, or title extracted from metadata
	Path            string // the full path
	RelativePath    string // the full path relative to the root media parent directory
	HaveMetadata    bool
	DurationString  string // extracted from ffmpeg output
	DurationSeconds int
	Tooltip         string
	InfoLink        string
	ModTime         time.Time
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func isDir(dir string) bool {
	stat, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func modTime(path string) time.Time {
	stat, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return stat.ModTime()
}

func formatDuration(ss int) string {
	// returns [[nd ]nh ] nm
	mm := ss / 60
	ss -= mm * 60
	if ss >= 30 {
		mm += 1
	}
	hh := mm / 60
	mm -= hh * 60
	dd := hh / 24
	hh -= dd * 24
	var rtn string
	if dd > 0 {
		rtn = fmt.Sprintf("%dd %dh %02dm", dd, hh, mm)
	} else if hh > 0 {
		rtn = fmt.Sprintf("%dh %02dm", hh, mm)
	} else {
		rtn = fmt.Sprintf("%dm", mm)
	}
	return rtn
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
	relativePath, err := filepath.Rel(root, parent)
	if err != nil {
		log.Fatal(err)
	}

	rtn := &MediaDirectory{
		Root:           root,
		Path:           parent, // will be initialised to a full path
		Leaf:           filepath.Base(parent),
		RelativePath:   relativePath,
		SubDirectories: make(map[string]*MediaDirectory, 0),
		Files:          make(map[string]*MediaFile, 0),
	}

	rtn.Refresh()

	if len(rtn.SubDirectories) == 0 && len(rtn.Files) == 0 {
		return nil // don't bother showing empty directories
	}

	return rtn
}

func (mediaDir *MediaDirectory) HasChanged() bool {
	if !isDir(mediaDir.Path) {
		return true // It's changed to nonexistent or not a directory
	}
	return mediaDir.ModTime.Compare(modTime(mediaDir.Path)) < 0
}

func (mediaDir *MediaDirectory) RefreshAndGetMetadata() {
	mediaDir.Refresh()
	go getMediaLengths(mediaDir)
}

func (mediaDir *MediaDirectory) Refresh() {
	// Note that this doesn't remove the directory from its parent if it's now empty
	mediaDir.ModTime = modTime(mediaDir.Path)

	files, err := os.ReadDir(mediaDir.Path)
	if err != nil {
		log.Println(err)
		mediaDir.SubDirectories = make(map[string]*MediaDirectory)
		mediaDir.Files = make(map[string]*MediaFile)
		return
	}

	subdirsDeleted := slices.Collect(maps.Keys(mediaDir.SubDirectories))
	filesDeleted := slices.Collect(maps.Keys(mediaDir.Files))

	for _, file := range files {
		fileName := file.Name()
		subPath := fmt.Sprintf("%s/%s", mediaDir.Path, fileName)
		if file.IsDir() {
			subdir, found := mediaDir.SubDirectories[fileName]
			if found {
				// directory already existed
				subdirsDeleted = slices.DeleteFunc(subdirsDeleted, func(e string) bool {
					return e == fileName
				})
				if subdir.HasChanged() {
					subdir.Refresh()
				}
			} else {
				// This is a new directory
				subdir = readMediaDir(mediaDir.Root, subPath)
				if subdir != nil {
					mediaDir.SubDirectories[file.Name()] = subdir
				}
			}
		} else {
			ext := filepath.Ext(fileName)
			if ext == ".mp3" || ext == ".m4a" {
				mediaFile, found := mediaDir.Files[fileName]
				if found {
					// file already existed
					filesDeleted = slices.DeleteFunc(filesDeleted, func(e string) bool {
						return e == fileName
					})
					if mediaFile.HasChanged() {
						mediaFile.ModTime = modTime(mediaFile.Path)
						mediaFile.HaveMetadata = false
					}
				} else {
					// new file
					displayName := strings.TrimSuffix(fileName, ext)
					relativePath := filepath.Join(mediaDir.RelativePath, fileName)
					mediaDir.Files[fileName] = &MediaFile{
						DisplayName:  displayName,
						Path:         subPath,
						RelativePath: relativePath,
						HaveMetadata: false,
						ModTime:      modTime(subPath),
					}
				}
			}
		}
	}

	for _, fileName := range subdirsDeleted {
		delete(mediaDir.SubDirectories, fileName)
	}
	for _, fileName := range filesDeleted {
		delete(mediaDir.Files, fileName)
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
	album := ""
	date := ""
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimLeft(line, " ")
		lineWords := strings.Fields(line)
		if len(lineWords) > 1 && lineWords[0] == "Duration:" {
			file.DurationString = strings.TrimSuffix(lineWords[1], ",")
			file.DurationSeconds = parseDurationString(file.DurationString)
			if i := strings.Index(file.DurationString, "."); i > 0 {
				file.DurationString = file.DurationString[:i]
			}
		} else if len(lineWords) > 2 && lineWords[0] == "title" {
			title := strings.TrimLeft(line[6:], " ")
			title = strings.TrimLeft(title[2:], " ")
			file.DisplayName = title
		} else if len(lineWords) > 2 && lineWords[0] == "album" {
			album = strings.TrimLeft(line[6:], " ")
			album = strings.TrimLeft(album[2:], " ")
		} else if len(lineWords) > 2 && lineWords[0] == "date" {
			date = lineWords[2]
			if t := strings.Index(date, "T"); t > -1 {
				date = date[:t]
			}
		} else if len(lineWords) >= 3 && lineWords[1] == "INFO:" && strings.HasPrefix(lineWords[2], "https://") {
			file.InfoLink = lineWords[2]
		}
	}
	file.Tooltip = buildTooltip(album, date)
	cmd.Wait()
	file.HaveMetadata = true
}

func buildTooltip(lines ...string) string {
	tooltip := ""
	for _, line := range lines {
		if line != "" {
			if tooltip != "" {
				tooltip += "<br>"
			}
			tooltip += html.EscapeString(line)
		}
	}
	return tooltip
}

func getMediaLengths(directory *MediaDirectory) {
	// handle the files in this directory
	// (so we populate the files in / first)
	for _, file := range directory.Files {
		if !file.HaveMetadata {
			getOneMediaInfo(file)
		}
	}

	// now recurse for subdirectories
	for _, subdir := range directory.SubDirectories {
		getMediaLengths(subdir)
	}

	// Calculate the total duration of the directory contents
	directory.TotalDurationSeconds = 0
	for _, subdir := range directory.SubDirectories {
		directory.TotalDurationSeconds += subdir.TotalDurationSeconds
	}
	for _, file := range directory.Files {
		directory.TotalDurationSeconds += file.DurationSeconds
	}
	directory.TotalDurationString = formatDuration(directory.TotalDurationSeconds)
}

func (media *RootMediaDirectory) MarkFilePlayed(file *MediaFile) error {
	if !isFile(file.Path) {
		return errors.New("file not found")
	}
	dest := filepath.Join(media.PlayedMediaDirectory, file.RelativePath)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	if err := os.Rename(file.Path, dest); err != nil {
		return err
	}

	// Remove the file from our records
	pathElts := make([]string, 0)
	filePath := file.RelativePath
	for filePath != "" {
		dir, leaf := filepath.Split(filePath)
		pathElts = append(pathElts, leaf)
		filePath = strings.TrimSuffix(dir, string(filepath.Separator))
	}
	if len(pathElts) == 0 {
		return errors.New("error splitting path " + file.RelativePath)
	}

	mediaDir := media.Contents
	for _, dir := range slices.Backward(pathElts[1:]) {
		mediaDir = mediaDir.SubDirectories[dir]
	}
	delete(mediaDir.Files, pathElts[0])
	return nil
}

func (mediaFile *MediaFile) HasChanged() bool {
	if !isFile(mediaFile.Path) {
		return true // It's changed to nonexistent or not a file
	}
	return mediaFile.ModTime.Compare(modTime(mediaFile.Path)) < 0
}

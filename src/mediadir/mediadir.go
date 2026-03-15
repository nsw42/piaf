package mediadir

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

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

	if rtn.IsEmpty() {
		return nil // don't bother showing empty directories
	}

	return rtn
}

func (mediaDir *MediaDirectory) HasChanged() bool {
	if !IsDir(mediaDir.Path) {
		return true // It's changed to nonexistent or not a directory
	}
	return mediaDir.ModTime.Compare(modTime(mediaDir.Path)) < 0
}

func (mediaDir *MediaDirectory) IsEmpty() bool {
	return len(mediaDir.SubDirectories) == 0 && len(mediaDir.Files) == 0
}

func (mediaDir *MediaDirectory) RefreshAndGetMetadata() {
	mediaDir.Refresh()
	go mediaDir.GetMediaLengths()
}

func (mediaDir *MediaDirectory) Refresh() {
	// Note that this doesn't remove the directory from its parent if it's now empty
	// but it does remove children that are now empty
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
			mediaDir.refreshSubDir(fileName, subPath, &subdirsDeleted)
		} else {
			mediaDir.refreshFile(fileName, subPath, &filesDeleted)
		}
	}

	for _, fileName := range subdirsDeleted {
		mediaDir.TotalDurationSeconds -= mediaDir.SubDirectories[fileName].TotalDurationSeconds
		delete(mediaDir.SubDirectories, fileName)
	}
	for _, fileName := range filesDeleted {
		mediaDir.TotalDurationSeconds -= mediaDir.Files[fileName].DurationSeconds
		delete(mediaDir.Files, fileName)
	}
}

func (mediaDir *MediaDirectory) refreshFile(fileName, subPath string, filesDeleted *[]string) {
	ext := filepath.Ext(fileName)
	if ext == ".mp3" || ext == ".m4a" {
		mediaFile, found := mediaDir.Files[fileName]
		if found {
			// file already existed
			*filesDeleted = slices.DeleteFunc(*filesDeleted, func(e string) bool {
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

func (mediaDir *MediaDirectory) refreshSubDir(fileName, subPath string, subdirsDeleted *[]string) {
	subdir, found := mediaDir.SubDirectories[fileName]
	if found {
		// directory already existed
		if subdir.HasChanged() {
			subdir.Refresh()
		}
		if subdir.IsEmpty() {
			// it already existed, but it's now empty - leave it in subdirsDeleted,
			// because there's no point showing an empty directory
		} else {
			// it already existed, and it's not empty
			*subdirsDeleted = slices.DeleteFunc(*subdirsDeleted, func(e string) bool {
				return e == fileName
			})
		}
	} else {
		// This is a new directory
		subdir = readMediaDir(mediaDir.Root, subPath)
		if subdir != nil {
			mediaDir.SubDirectories[fileName] = subdir
		}
	}
}

func (mediaDir *MediaDirectory) GetMediaLengths() {
	// handle the files in this directory
	// (so we populate the files in / first)
	for _, file := range mediaDir.Files {
		if !file.HaveMetadata {
			file.UpdateMetadata()
		}
	}

	// now recurse for subdirectories
	for _, subdir := range mediaDir.SubDirectories {
		subdir.GetMediaLengths()
	}

	// Calculate the total duration of the directory contents
	mediaDir.TotalDurationSeconds = 0
	for _, subdir := range mediaDir.SubDirectories {
		mediaDir.TotalDurationSeconds += subdir.TotalDurationSeconds
	}
	for _, file := range mediaDir.Files {
		mediaDir.TotalDurationSeconds += file.DurationSeconds
	}
	mediaDir.TotalDurationString = formatDuration(mediaDir.TotalDurationSeconds)
}

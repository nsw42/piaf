package mediadir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type RootMediaDirectory struct {
	RootMediaDirectory     string
	UnplayedMediaDirectory string
	PlayedMediaDirectory   string
	Contents               *MediaDirectory
}

func ReadMediaDir(root string) (*RootMediaDirectory, error) {
	media := &RootMediaDirectory{
		RootMediaDirectory:     root,
		UnplayedMediaDirectory: filepath.Join(root, "Unplayed"),
		PlayedMediaDirectory:   filepath.Join(root, "Played"),
	}

	if !IsDir(media.RootMediaDirectory) {
		return nil, fmt.Errorf("root media directory '%s' does not exist", root)
	}
	if !IsDir(media.UnplayedMediaDirectory) {
		return nil, fmt.Errorf("unplayed directory '%s' does not exist", root)
	}
	if !IsDir(media.PlayedMediaDirectory) {
		return nil, fmt.Errorf("played directory '%s' does not exist", root)
	}

	media.Contents = readMediaDir(media.UnplayedMediaDirectory, media.UnplayedMediaDirectory)
	go media.Contents.GetMediaLengths()

	return media, nil
}

func (root *RootMediaDirectory) MarkFilePlayed(file *MediaFile) error {
	if !IsFile(file.Path) {
		return errors.New("file not found")
	}
	dest := filepath.Join(root.PlayedMediaDirectory, file.RelativePath)
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

	mediaDir := root.Contents
	mediaDir.TotalDurationSeconds -= file.DurationSeconds
	for _, dir := range slices.Backward(pathElts[1:]) {
		mediaDir = mediaDir.SubDirectories[dir]
		mediaDir.TotalDurationSeconds -= file.DurationSeconds
	}
	delete(mediaDir.Files, pathElts[0])
	return nil
}

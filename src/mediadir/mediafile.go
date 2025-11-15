package mediadir

import (
	"html"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

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

func (mediaFile *MediaFile) HasChanged() bool {
	if !IsFile(mediaFile.Path) {
		return true // It's changed to nonexistent or not a file
	}
	return mediaFile.ModTime.Compare(modTime(mediaFile.Path)) < 0
}

func (file *MediaFile) UpdateMetadata() {
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
		file.processOneMetadataLine(line, &album, &date)
	}
	file.Tooltip = buildTooltip(album, date)
	cmd.Wait()
	file.HaveMetadata = true
}

func (file *MediaFile) processOneMetadataLine(line string, album *string, date *string) {
	line = strings.TrimLeft(line, " ")
	lineWords := strings.Fields(line)
	if len(lineWords) > 1 && lineWords[0] == "Duration:" {
		file.DurationString = strings.TrimSuffix(lineWords[1], ",")
		file.DurationSeconds = parseDurationString(file.DurationString)
		if i := strings.Index(file.DurationString, "."); i > 0 {
			file.DurationString = file.DurationString[:i]
		}
	} else if len(lineWords) > 2 && lineWords[0] == "title" && file.DisplayName == "" {
		title := strings.TrimLeft(line[6:], " ")
		title = strings.TrimLeft(title[2:], " ")
		file.DisplayName = title
	} else if len(lineWords) > 2 && lineWords[0] == "album" {
		*album = strings.TrimLeft(line[6:], " ")
		*album = strings.TrimLeft((*album)[2:], " ")
	} else if len(lineWords) > 2 && lineWords[0] == "date" {
		*date = lineWords[2]
		if t := strings.Index(*date, "T"); t > -1 {
			*date = (*date)[:t]
		}
	} else if len(lineWords) >= 3 && lineWords[1] == "INFO:" && strings.HasPrefix(lineWords[2], "https://") {
		file.InfoLink = lineWords[2]
	}
}

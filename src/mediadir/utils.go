package mediadir

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func IsDir(dir string) bool {
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

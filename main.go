package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	spaceRegex = regexp.MustCompile(`\s+`)
)

type window struct {
	title    string
	windowId int
	desktop  int
	pid      int
}

type timeSlice struct {
	start time.Time
	end   time.Time
}

type tracker struct {
	lastTitle     string
	idToTimeSlice map[string][]timeSlice
}

func (t *tracker) calculate() map[string]time.Duration {
	m := make(map[string]time.Duration)
	for k, v := range t.idToTimeSlice {
		for _, tc := range v {
			m[k] += tc.end.Sub(tc.start)
		}
	}

	return m
}

func (t *tracker) update(title string) {
	if t.idToTimeSlice == nil {
		t.idToTimeSlice = make(map[string][]timeSlice)
	}
	if t.lastTitle == title {
		l := len(t.idToTimeSlice[title])
		t.idToTimeSlice[title][l-1].end = time.Now()

		return
	}

	_, ok := t.idToTimeSlice[t.lastTitle]

	if ok {
		l := len(t.idToTimeSlice[t.lastTitle])
		t.idToTimeSlice[t.lastTitle][l-1].end = time.Now()
	}

	t.idToTimeSlice[title] = append(t.idToTimeSlice[title], timeSlice{time.Now(), time.Now()})

	t.lastTitle = title
}

func parseWindow(line string) *window {
	fields := spaceRegex.Split(line, 5)
	winId, _ := strconv.ParseInt(strings.TrimPrefix(fields[0], "0x"), 16, 64)
	desktopId, _ := strconv.Atoi(fields[1])
	pid, _ := strconv.Atoi(fields[2])

	return &window{
		title:    fields[4],
		windowId: int(winId),
		desktop:  desktopId,
		pid:      pid,
	}
}

func getWindows() []*window {
	cmd := exec.Command("wmctrl", "-lp")
	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
	windows := make([]*window, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		win := parseWindow(line)
		windows = append(windows, win)
	}

	return windows
}

func getActiveWinTitle() string {
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")

	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	out, err := cmd.Output()

	// sometimes fails, ignore
	if err != nil {
		// log.Fatal(err.Error() + stderr.String())
		return ""
	}

	return strings.TrimSpace(string(out))
}

func main() {
	var t tracker
	for range time.Tick(500 * time.Millisecond) {
		title := getActiveWinTitle()
		if title == "" {
			continue
		}

		t.update(title)
		// pp.Println(t.calculate())

		vals := t.calculate()
		keys := make([]string, 0, len(vals))
		for k := range vals {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		fmt.Println(strings.Repeat(">", 20))
		for _, k := range keys {
			fmt.Printf("\r%20s: %s\n", vals[k].String(), k)
		}
		fmt.Println(strings.Repeat("<", 20))
	}
}

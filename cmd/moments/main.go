package main

import (
	"flag"
	"fmt"
	"log"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/joho/godotenv"
)

var (
	pause  = 10
	suffix = `_record.json`

	DIR = flag.String(
		"d",
		`./assets/`,
		`-d="./assets/"`,
	)

	LOG = flag.Bool(
		"l",
		false,
		"-l",
	)
	RECORD = flag.Bool(
		"r",
		false,
		"-r",
	)
	PLAY = flag.Bool(
		"p",
		false,
		"-p",
	)
	LOOP = flag.Bool(
		"i",
		false,
		"-i",
	)
	BACKWARDS = flag.Bool(
		"b",
		false,
		"-b",
	)
	FILTER = flag.String(
		"t",
		"",
		`-t "firefox"`,
	)
	FILE = flag.String(
		"f",
		"",
		`-f "the_one_about_the_duck.json"`,
	)

	help_msg = `moments - record and replay moments at your computer
	
usage: moments [-flag, ...]

				flags:
				
	-b	Sort moments backwards 
	-d 	Set moments directory, eg -d "./assets/"
	-f 	Set moment name, eg -f "record.json"
	-i	Set moment to infitite loop 
	-l	Logs moments to the console
	-p	Play moments, default latest, HANDLE WITH CARE  
	-r	Record moments, eg. window title and events from input devices 
	-t 	Filter moments by window title, eg -t "firefox"

				examples:

logging only: 
	moments -l 					; log moments to the console
	
recording:
	moments -r 					; record moments
	moments -r -l 					; record & log moments to the console
	moments -r -d "./assets/" -l			; record moments to specified directory & print logs the console
	moments -r -f "simple.json"	-l 		; record to "simple.json" file and log to console
	moments -r -d "./assets/" -f "simple.json" -l 	; record "simple.json" to specified dir and log to console
	moments -r -t "firefox" -l			; record moments when window title is "firefox" and log to console

playback:
	moments -p 					; play from default dir 
	moments -p -l 					; play from default dir and log moments to the console
	moments -p -d "./assets/"			; play latest timestamp filename in the directory
	moments -p -f "simple.json" 			; play specified file
	moments -p -i -f "simple.json" 			; play on infinite loop, specified file  
	moments -p -b -i -f "simple.json" 		; play backwards, on infinite loop, plays specified file 

`
)

func main() {

	robotgo.MouseSleep = pause
	robotgo.KeySleep = robotgo.MouseSleep

	godotenv.Load()

	flag.Parse()
	flag_count := flag.NFlag()

	if flag_count <= 0 {
		fmt.Print(help_msg)
	} else {
		var values []string
		flag.VisitAll(
			func(f *flag.Flag) {
				if f.Name == "p" || f.Name == "r" || f.Name == "l" {
					values = append(values, f.Value.String())
				}
			},
		)

		file_path := resolve_filename(*FILE)

		if *FILTER != "" && !*PLAY && !*RECORD {
			for _, moment := range Filter(
				file_path,
				*FILTER,
			) {
				fmt.Printf(
					"%s\n",
					moment,
				)
			}
		} else if len(values) < 3 ||
			!*PLAY && !*RECORD && !*LOG ||
			*PLAY && *RECORD {
			panic("moments can either log or not when moments is set to play or record")
		}

		if *LOG && !*PLAY && !*RECORD {
			Listen(
				*FILTER,
				*LOG,
			)
		}

		if *RECORD || *RECORD && *LOG {
			Record(
				file_path,
				Listen(
					*FILTER,
					*LOG,
				),
			)
		} else if *PLAY || *PLAY && *LOG {
			Play(
				file_path,
				*FILTER,
				*BACKWARDS,
				*LOOP,
				*LOG,
			)
		} else if *FILTER != "" && !*PLAY && !*RECORD {

		}
	}
}

func resolve_filename(file string) string {
	var err error
	if *FILE != "" {

		return *DIR + file

	} else if *RECORD {
		return time_stamp_filename()

	} else {
		file, err = newest_recording(*DIR)
		if err != nil {
			log.Fatalf("Error finding most recent file: %v", err)
		}
	}
	return file
}

func time_stamp_filename() string {
	now := time.Now().Local().Format("2006_01_02_150405")
	return *DIR + now + suffix
}

// Function to find the most recent file based on the timestamp in the filename
func newest_recording(dir string) (string, error) {

	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	var latest string
	var current time.Time // default is 00:00:00

	for _, file := range files {
		if file.IsDir() ||
			file.Name() == ".keep" ||
			!strings.Contains(file.Name(), suffix) {
			continue
		}

		time_stamp := strings.TrimSuffix(file.Name(), suffix)

		timestamp, err := time.Parse(
			"2006_01_02_150405",
			time_stamp,
		)
		if err != nil {
			return "", fmt.Errorf("error parsing timestamp: %v", err)
		}

		// Determine if this file is more current
		if timestamp.After(current) {
			current = timestamp
			latest = file.Name()
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no files found")
	}

	return filepath.Join(dir, latest), nil
}

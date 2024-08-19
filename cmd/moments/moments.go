package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type (
	moment struct {
		Title string     `json:"title"`
		Event hook.Event `json:"event"`
	}
	moments []moment
)

func Listen(filter string, LOG bool) moments {

	// Create a channel to handle signals
	listener := make(chan os.Signal, 1)
	signal.Notify(
		listener,
		syscall.SIGINT,
		syscall.SIGTERM,
		// we could listen for other system calls
	)

	// Create a slice to accumulate moments
	var moments moments

	// Mutex to protect access to the moments slice
	var semaphor sync.Mutex

	// Create a channel to signal termination of the event loop
	done := make(chan struct{})

	go func() {
		// Start capturing events
		events := hook.Start()

		// then range those events as we gather them
		for event := range events {
			// get the window title of the moment and lowercase the string
			window_title := strings.ToLower(robotgo.GetTitle())

			// determine if we are filtering for term
			if filter != "" {
				if !strings.Contains(
					window_title,
					strings.ToLower(filter),
				) {
					continue
				}
			}

			// first lock the slice until this loop is done
			semaphor.Lock()

			// build the moment
			moment := moment{
				Title: window_title,
				Event: event,
			}

			if LOG {
				fmt.Printf("%+v\n", moment)
			}

			// add moment to the collection of moments
			moments = append(
				moments,
				moment,
			)

			// no unlock
			semaphor.Unlock()
		}
		done <- struct{}{} // Notify that event processing is done
	}()

	// Wait for an interrupt signal
	<-listener
	fmt.Println("Received interrupt signal, stopping...")

	// Close the event loop
	hook.End()

	// Wait for the event loop to finish
	<-done

	return moments
}

func Record(filename string, moments moments) {

	var semaphor sync.Mutex

	// Open the file in write mode, or create it if it doesn't exist
	file, err := os.OpenFile(
		filename,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0600,
	)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed when the function returns

	// Write the JSON data to the file
	semaphor.Lock()
	// fmt.Printf("%+v", moments)
	momentsJSON, err := json.MarshalIndent(
		moments,
		"",
		"	",
	)
	semaphor.Unlock()

	if err != nil {
		fmt.Println("Error marshaling moments to JSON:", err)
		return
	}

	if _, err := file.Write(momentsJSON); err != nil {
		fmt.Println("Error writing to file:", err)
	}

}
func Play(file_path, filter_term string, REVERSE, LOOP, LOG bool) {
	// Unmarshal the JSON data into a slice of Event
	var moments moments = Filter(file_path, filter_term)

	if REVERSE {
		reverse(moments)
	}

	if !LOOP {
		process_moments(moments, LOG)
	} else {
		for {
			process_moments(moments, LOG)
		}
	}
}
func Filter(file_path, filter_term string) moments {
	var (
		filtered, moments moments
	)
	if err := json.Unmarshal(
		get_data(file_path),
		&moments,
	); err != nil {
		panic(err)
	}

	for _, moment := range moments {
		if filter_term != `` {
			condition := strings.Contains(
				strings.ToLower(moment.Title),
				strings.ToLower(filter_term),
			)
			if !condition {
				continue
			}
		}
		filtered = append(filtered, moment)
	}
	return filtered
}
func get_data(file_path string) []byte {
	// Open the JSON file for reading
	file, err := os.Open(file_path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Read the entire file content
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	return data
}

func reverse(moments moments) {
	sort.Slice(
		moments,
		func(i, j int) bool {
			return moments[i].Event.When.After(
				moments[j].Event.When,
			)
		},
	)
}

func process_moments(moments moments, LOG bool) {
	for _, moment := range moments {

		if LOG {
			log.Printf("%+v\n", moment)
		}
		switch moment.Event.Kind {
		// case hook.MouseDrag: // not implemented very well
		// 	robotgo.DragSmooth(
		// 		int(moment.Event.X),
		// 		int(moment.Event.Y),
		// 	)
		case hook.MouseDown,
			hook.MouseUp,
			hook.MouseMove:

			robotgo.Move(
				int(moment.Event.X),
				int(moment.Event.Y),
			)

			if moment.Event.Clicks > 0 {
				robotgo.Click()
			}

		case hook.KeyDown,
			hook.KeyUp:

			key := string(moment.Event.Keychar)

			if moment.Event.Kind == hook.KeyDown ||
				moment.Event.Keychar != hook.KeyHold {

				robotgo.KeyTap(key)

			} else {
				// Implement KeyUp if needed by robotgo, or handle differently
			}
		}
	}
}

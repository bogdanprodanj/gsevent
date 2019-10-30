package file

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	eventFileNameFormat = "events_%d"
	eventFolder         = "events"
)

// EventRecorder is responsible for writing event data into CSV file
type EventRecorder struct {
	events       chan []string
	quit         chan struct{}
	fileInterval time.Duration
	w            io.WriteCloser
}

// NewEventRecorder returns new event file recorder by creating folder and initial file for storing events
func NewEventRecorder(events chan []string, fileInterval time.Duration) (*EventRecorder, error) {
	var err error
	if err = os.MkdirAll(eventFolder, 0755); err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		return nil, err
	}
	var f *os.File
	if f, err = os.Create(filepath.Join(eventFolder, fmt.Sprintf(eventFileNameFormat, time.Now().UnixNano()))); err != nil {
		return nil, err
	}
	return &EventRecorder{
		events:       events,
		quit:         make(chan struct{}),
		fileInterval: fileInterval,
		w:            f,
	}, nil
}

// Start start to consume events from the event channel and writes them into file.
// To avoid extra large CSV file, recorder will create new files with interval (defaults to 1 hour) provided in configs.
func (er *EventRecorder) Start() {
	var err error
	w := csv.NewWriter(er.w)
	interval := time.NewTicker(er.fileInterval)
loop:
	for {
		select {
		case e, ok := <-er.events:
			if !ok {
				break loop
			}
			if err = w.Write(e); err != nil {
				log.WithError(err).Errorf("failed to record event %+v", e)
			}
		case <-interval.C:
			if er.w, err = os.Create(filepath.Join(eventFolder, fmt.Sprintf(eventFileNameFormat, time.Now().UnixNano()))); err != nil {
				log.WithError(err).Fatal("failed to create file for recording events")
			}
			w = csv.NewWriter(er.w)
		case <-er.quit:
			interval.Stop()
			w.Flush()
			_ = er.w.Close()
			break loop
		}
	}
}

// Stop stops event recorder by closing quit channel
func (er *EventRecorder) Stop() {
	close(er.quit)
}

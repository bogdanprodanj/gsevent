package storage

import "github.com/gsevent/runtime/models"

type Cacher interface {
	AddEvent(e *models.Event) error
	ListEvents() ([]string, error)
	EventCount(eventType string, start, end int) (int64, error)
	EventData(eventType string, start, end int) ([]models.Data, error)
	Stop()
}

type Recorder interface {
	Start()
	Stop()
}

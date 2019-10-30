package runtime

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/friendsofgo/errors"
	"github.com/gammazero/workerpool"
	"github.com/gsevent/runtime/storage"
	"github.com/gsevent/runtime/storage/file"
	"github.com/gsevent/runtime/storage/redis"
)

// this value might be tweaked according writes capabilities of the Recorder
const eventsChanSize = 10000

// App is core element of this application and contains all the necessary dependencies
type App struct {
	Router   *fasthttprouter.Router
	redis    storage.Cacher
	recorder storage.Recorder
	events   chan []string
	wp       *workerpool.WorkerPool
}

// NewApp creates new App with file recorder, redis client and worker pool,
// which manages events sending to the file recorder
func NewApp(cfg *Config) (*App, error) {
	events := make(chan []string, eventsChanSize)
	recorder, err := file.NewEventRecorder(events, cfg.NewFileInterval)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create event recorder")
	}
	cache, err := redis.NewRedis(cfg.RedisAddress, cfg.RedisPassword)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create redis client")
	}
	return &App{
		Router:   fasthttprouter.New(),
		events:   events,
		redis:    cache,
		recorder: recorder,
		wp:       workerpool.New(cfg.MaxWorkers),
	}, nil
}

// Start inits the routes and starts file recorder
func (a *App) Start() {
	a.registerRoutes()
	go a.recorder.Start()
}

// Stop waits for all the workers to finish sending events to the file recorder while not
// accepting new events, stops file recorder and redis client
func (a *App) Stop() {
	a.wp.StopWait()
	a.recorder.Stop()
	a.redis.Stop()
}

func (a *App) registerRoutes() {
	a.Router.POST("/events", a.AddEvent)
	a.Router.GET("/events/:event_type", a.GetEventData)
	a.Router.GET("/events/:event_type/count", a.EventCount)
}

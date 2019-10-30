package runtime

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gsevent/runtime/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const (
	startQueryParam    = "start"
	endQueryParam      = "end"
	eventTypePathParam = "event_type"
)

// AddEvent adds incoming event to redis and submits this event to a worker for
// further processing by file recorder
func (a *App) AddEvent(ctx *fasthttp.RequestCtx) {
	e := new(models.Event)
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(ctx.Request.Body(), e)
	if err != nil {
		ctx.Error(errors.WithMessage(err, "invalid input").Error(), http.StatusBadRequest)
		return
	}
	if err = e.Validate(nil); err != nil {
		ctx.Error(errors.WithMessage(err, "invalid input").Error(), http.StatusBadRequest)
		return
	}
	if err = a.redis.AddEvent(e); err != nil {
		logrus.WithError(err).Error("failed to save event into cache")
		ctx.Error("failed to save event", http.StatusInternalServerError)
		return
	}
	a.wp.Submit(func() {
		b, _ := jsoniter.MarshalToString(e.Params)
		a.events <- []string{e.EventType, strconv.Itoa(int(*e.Ts)), b}
	})
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

// GetEventData returns list of data for provided event type and time range
func (a *App) GetEventData(ctx *fasthttp.RequestCtx) {
	end := ctx.QueryArgs().GetUintOrZero(endQueryParam)
	if end == 0 {
		end = int(time.Now().Unix())
	}
	data, err := a.redis.EventData(fmt.Sprint(ctx.UserValue(eventTypePathParam)), ctx.QueryArgs().GetUintOrZero(startQueryParam), end)
	if err != nil {
		logrus.WithError(err).Error("failed to get events data from cache")
		ctx.Error("failed to get events data", http.StatusInternalServerError)
		return
	}
	body, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("failed to get marshal events data")
		ctx.Error("failed to get events data", http.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(body)
}

// EventCount returns number of events for provided event type and time range
func (a *App) EventCount(ctx *fasthttp.RequestCtx) {
	end := ctx.QueryArgs().GetUintOrZero(endQueryParam)
	if end == 0 {
		end = int(time.Now().Unix())
	}
	count, err := a.redis.EventCount(fmt.Sprint(ctx.UserValue(eventTypePathParam)), ctx.QueryArgs().GetUintOrZero(startQueryParam), end)
	if err != nil {
		logrus.WithError(err).Error("failed to get events number from cache")
		ctx.Error("failed to get events number", http.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(strconv.Itoa(int(count)))
}

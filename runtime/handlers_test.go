package runtime

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/golang/mock/gomock"
	"github.com/gsevent/mocks"
	"github.com/gsevent/runtime/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestAddEvent(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockCacher(ctrl)
	events := make(chan []string, 2)
	app := App{
		redis:  mockRedis,
		wp:     workerpool.New(1),
		events: events,
	}
	t.Run("happy path", func(t *testing.T) {
		mockRedis.EXPECT().AddEvent(gomock.Any()).Return(nil)
		requestCtx := &fasthttp.RequestCtx{}
		now := time.Now().Unix()
		e := &models.Event{
			Data: models.Data{
				Params: map[string]interface{}{"user": "hello"},
				Ts:     &now,
			},
			EventType: "click",
		}
		b, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(e)
		requestCtx.Request.SetBody(b)
		app.AddEvent(requestCtx)
		assert.Equal(t, requestCtx.Response.StatusCode(), fasthttp.StatusAccepted)
		app.wp.StopWait()
		select {
		case e := <-events:
			assert.Equal(t, e, []string{"click", strconv.Itoa(int(now)), `{"user":"hello"}`})
		case <-time.NewTimer(time.Second).C:
			t.Error("event should be send over a channel")
		}
	})
	t.Run("invalid request", func(t *testing.T) {
		requestCtx := &fasthttp.RequestCtx{}
		e := &models.Event{}
		b, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(e)
		requestCtx.Request.SetBody(b)
		app.AddEvent(requestCtx)
		assert.Equal(t, requestCtx.Response.StatusCode(), fasthttp.StatusBadRequest)
		assert.Contains(t, string(requestCtx.Response.Body()), "ts in body is required")
	})
	t.Run("invalid request body format", func(t *testing.T) {
		requestCtx := &fasthttp.RequestCtx{}
		requestCtx.Request.SetBody([]byte(".."))
		app.AddEvent(requestCtx)
		assert.Equal(t, requestCtx.Response.StatusCode(), fasthttp.StatusBadRequest)
		assert.Contains(t, string(requestCtx.Response.Body()), "invalid input")
	})
	t.Run("cache internal error", func(t *testing.T) {
		mockRedis.EXPECT().AddEvent(gomock.Any()).Return(errors.New("some error"))
		requestCtx := &fasthttp.RequestCtx{}
		now := time.Now().Unix()
		e := &models.Event{
			Data: models.Data{
				Params: map[string]interface{}{"user": "hello"},
				Ts:     &now,
			},
			EventType: "click",
		}
		b, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(e)
		requestCtx.Request.SetBody(b)
		app.AddEvent(requestCtx)
		assert.Equal(t, requestCtx.Response.StatusCode(), fasthttp.StatusInternalServerError)
	})
}

func TestListEvents(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockCacher(ctrl)
	app := App{redis: mockRedis}
	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.SetRequestURI("/events")
	t.Run("happy path", func(t *testing.T) {
		mockRedis.EXPECT().ListEvents().Return([]string{"foo", "bar"}, nil)
		app.ListEvents(requestCtx)
		assert.Equal(t, string(requestCtx.Response.Body()), `["foo","bar"]`)
	})
	t.Run("cache internal error", func(t *testing.T) {
		mockRedis.EXPECT().ListEvents().Return(nil, errors.New("some error"))
		app.ListEvents(requestCtx)
		assert.Equal(t, requestCtx.Response.StatusCode(), fasthttp.StatusInternalServerError)
	})
}

func TestEventCount(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockCacher(ctrl)
	app := App{redis: mockRedis}
	mockRedis.EXPECT().EventCount(gomock.Any(), 0, gomock.Any()).Return(int64(10), nil)
	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.SetRequestURI("/events/hello/count")
	app.EventCount(requestCtx)
	assert.Equal(t, string(requestCtx.Response.Body()), "10")
}

func TestGetEventData(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRedis := mocks.NewMockCacher(ctrl)
	app := App{redis: mockRedis}
	ts := time.Now().Unix()
	mockRedis.EXPECT().EventData(gomock.Any(), 0, gomock.Any()).Return([]models.Data{{
		Params: map[string]interface{}{"name": "foo"},
		Ts:     &ts,
	}}, nil)
	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.Request.SetRequestURI("/events/hello")
	app.EventData(requestCtx)
	assert.Equal(t, string(requestCtx.Response.Body()), fmt.Sprintf(`[{"params":{"name":"foo"},"ts":%d}]`, ts))
}

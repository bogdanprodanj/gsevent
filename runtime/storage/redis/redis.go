package redis

import (
	"io"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/gsevent/runtime/models"
)

// Redis is a wrapper for redis client
type Redis struct {
	client *redis.Client
	io.ReadCloser
}

// NewRedis creates new Redis and pings client to confirm successful connection
func NewRedis(redisAddress, redisPassword string) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

// AddEvent adds new event to redis
func (r *Redis) AddEvent(e *models.Event) error {
	return r.client.ZAdd(e.EventType, &redis.Z{
		Score:  float64(*e.Ts),
		Member: &e.Data,
	}).Err()
}

// ListEvents returns list of existing events types
func (r *Redis) ListEvents() ([]string, error) {
	return r.client.Keys("*").Result()
}

// EventData returns list of data for provided event type and time range
func (r *Redis) EventData(eventType string, start, end int) ([]models.Data, error) {
	var data []models.Data
	err := r.client.ZRangeByScore(eventType, &redis.ZRangeBy{
		Min: strconv.Itoa(start),
		Max: strconv.Itoa(end),
	}).ScanSlice(&data)
	return data, err
}

// EventCount returns number of events for provided event type and time range
func (r *Redis) EventCount(eventType string, start, end int) (int64, error) {
	return r.client.ZCount(eventType, strconv.Itoa(start), strconv.Itoa(end)).Result()
}

// Stops closes redis client
func (r *Redis) Stop() {
	_ = r.client.Close()
}

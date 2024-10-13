package query

import (
	"errors"
	"strings"
	"time"

	"github.com/umgbhalla/gokv/internal/store"
)

type Query struct {
	store *store.Store
}

func New(s *store.Store) *Query {
	return &Query{store: s}
}

func (q *Query) Execute(queryString string) (interface{}, error) {
	parts := strings.Fields(queryString)
	if len(parts) == 0 {
		return nil, errors.New("empty query")
	}

	command := strings.ToUpper(parts[0])
	switch command {
	case "GET":
		if len(parts) != 2 {
			return nil, errors.New("GET query should have exactly one argument")
		}
		return q.executeGet(parts[1])
	case "SET":
		if len(parts) < 3 {
			return nil, errors.New("SET query should have at least two arguments")
		}
		ttl := 24 * time.Hour
		if len(parts) == 4 {
			duration, err := time.ParseDuration(parts[3])
			if err != nil {
				return nil, errors.New("invalid TTL format")
			}
			ttl = duration
		}
		return nil, q.executeSet(parts[1], parts[2], ttl)
	case "DELETE":
		if len(parts) != 2 {
			return nil, errors.New("DELETE query should have exactly one argument")
		}
		return nil, q.executeDelete(parts[1])
	case "SCAN":
		if len(parts) != 2 {
			return nil, errors.New("SCAN query should have exactly one argument")
		}
		return q.executeScan(parts[1])
	default:
		return nil, errors.New("unknown command")
	}
}

func (q *Query) executeGet(key string) (interface{}, error) {
	value, exists := q.store.Get(key)
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (q *Query) executeSet(key string, value interface{}, ttl time.Duration) error {
	q.store.Set(key, value, ttl)
	return nil
}

func (q *Query) executeDelete(key string) error {
	q.store.Delete(key)
	return nil
}

// TODO: find faster mech for this ?
func (q *Query) executeScan(prefix string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	allData := q.store.GetAll()
	for k, v := range allData {
		if strings.HasPrefix(k, prefix) {
			result[k] = v.Data
		}
	}
	return result, nil
}

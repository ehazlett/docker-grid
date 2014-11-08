package datastore

import (
	"sync"
	"time"
)

type (
	Item struct {
		sync.RWMutex
		expires *time.Time
		Data    interface{}
	}
)

func (item *Item) expired() bool {
	expired := false
	item.RLock()
	if item.expires == nil {
		return true
	}
	expired = item.expires.Before(time.Now())
	item.RUnlock()
	return expired
}

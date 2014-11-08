package datastore

import (
	"errors"
	"sync"
	"time"
)

type (
	Datastore struct {
		mutex sync.RWMutex
		data  map[string]*Item
		ttl   time.Duration
	}
)

var (
	ErrKeyDoesNotExist = errors.New("key does not exist")
)

func New(ttl time.Duration) (*Datastore, error) {
	d := &Datastore{
		ttl:  ttl,
		data: map[string]*Item{},
	}

	ticker := time.NewTicker(ttl)

	go func() {
		for _ = range ticker.C {
			d.cleanup()
		}
	}()

	return d, nil
}

func (d *Datastore) Items() map[string]*Item {
	return d.data
}

func (d *Datastore) Set(key string, data interface{}) error {
	exp := time.Now().Add(d.ttl)
	item := &Item{
		expires: &exp,
		Data:    data,
	}
	d.data[key] = item
	return nil
}

func (d *Datastore) Get(key string) (interface{}, error) {
	if v, ok := d.data[key]; ok {
		return v, nil
	}
	return nil, ErrKeyDoesNotExist
}

func (d *Datastore) cleanup() {
	d.mutex.Lock()
	for key, item := range d.data {
		if item.expired() {
			delete(d.data, key)
		}
	}
	d.mutex.Unlock()
}

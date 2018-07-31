package inmem

import (
	"fmt"
	"log"
	"sync"

	"github.com/asnelzin/stream-api/pkg/store"
	"time"

	"context"
	"github.com/google/uuid"
)

type inmemStore struct {
	lock sync.RWMutex
	data map[string]store.Stream

	finishAfter int
	timers      map[string]*time.Timer
}

func NewStore(finishAfter int) store.Engine {
	return &inmemStore{
		data:        map[string]store.Stream{},
		finishAfter: finishAfter,
		timers:      map[string]*time.Timer{},
	}
}

func (m *inmemStore) Create() (*store.Stream, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	id := uuid.New().String()
	stream := store.Stream{
		ID:      id,
		State:   store.CREATED,
		Created: time.Now().UTC().Format(time.RFC3339Nano),
	}
	log.Printf("[INFO] save stream %v", stream)
	m.data[id] = stream

	return &stream, nil
}

func (m *inmemStore) List() ([]*store.Stream, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var streams []*store.Stream

	for _, s := range m.data {
		ls := s
		streams = append(streams, &ls)
	}
	return streams, nil
}

func (m *inmemStore) Start(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	s, ok := m.data[id]
	if !ok {
		return fmt.Errorf("could not find stream with id %s", id)
	}
	if s.State == store.FINISHED {
		return fmt.Errorf("could not change state of stream %s (stream is already finished)", id)
	}

	if s.State == store.INTERRUPTED {
		timer, ok := m.timers[id]
		if !ok {
			// sanity check
			return fmt.Errorf("could not change state of stream %s (stream is interrupted but no way to cacnel finish)", id)
		}
		timer.Stop()
		delete(m.timers, id)
	}

	s.State = store.ACTIVE
	m.data[id] = s

	return nil
}

func (m *inmemStore) Interrupt(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	s, ok := m.data[id]
	if !ok {
		return fmt.Errorf("could not find stream with id %s", id)
	}

	if s.State != store.ACTIVE {
		return fmt.Errorf("could not change state of stream %s (stream is not active)", id)
	}

	s.State = store.INTERRUPTED
	m.data[id] = s

	// start timer for finishing stream
	timer := time.AfterFunc(time.Duration(m.finishAfter)*time.Second, func() {
		m.finish(id)
	})
	m.timers[id] = timer
	return nil
}

func (m *inmemStore) startTimer(ctx context.Context, id string) {
	select {
	case <-time.After(time.Duration(m.finishAfter) * time.Second):
		m.finish(id)
	case <-ctx.Done():
		return
	}
}

func (m *inmemStore) finish(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	stream, ok := m.data[id]
	if !ok {
		log.Printf("[DEBUG] could not finish stream %s: deleted", id)
		return
	}

	if stream.State != store.INTERRUPTED {
		log.Printf("[DEBUG] could not finish stream %v: not interrupted", stream)
		return
	}

	stream.State = store.FINISHED
	m.data[id] = stream
}

func (m *inmemStore) Delete(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	s, ok := m.data[id]
	if !ok {
		return fmt.Errorf("could not find stream with id %s", id)
	}

	// delete should clear timer for object
	if s.State == store.INTERRUPTED {
		timer, ok := m.timers[id]
		if !ok {
			// sanity check
			return fmt.Errorf("could not change state of stream %s (stream is interrupted but no way to cacnel finish)", id)
		}
		timer.Stop()
		delete(m.timers, id)
	}

	delete(m.data, id)
	return nil
}

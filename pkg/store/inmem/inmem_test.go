package inmem

import (
	"github.com/asnelzin/stream-api/pkg/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestInmemStore_Create(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	assert.NoError(t, err)

	// ensure ID is a valid UUID
	_, err = uuid.Parse(stream.ID)
	assert.Nil(t, err)
	assert.Equal(t, store.CREATED, stream.State)
}

func TestInmemStore_Start(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	err = mem.Start(stream.ID)
	assert.NoError(t, err)

	streams, err := mem.List()
	require.NoError(t, err)
	assert.Equal(t, store.ACTIVE, streams[0].State)
}

func TestInmemStore_Start_NotFound(t *testing.T) {
	mem := prep(t)

	err := mem.Start("8dff7c72-3edb-4718-87e9-6d60f653b4cf")
	assert.Error(t, err)
}

func TestInmemStore_Start_Finished(t *testing.T) {
	mem := prep(t)

	stream, _ := mem.Create()
	mem.Start(stream.ID)
	mem.Interrupt(stream.ID)

	// more than 1s because scheduler should have time to switch
	time.Sleep(1100 * time.Millisecond)

	err := mem.Start(stream.ID)
	assert.Error(t, err)
}

func TestInmemStore_Interrupt(t *testing.T) {
	mem := prep(t)

	stream, _ := mem.Create()
	mem.Start(stream.ID)

	err := mem.Interrupt(stream.ID)
	assert.NoError(t, err)

	streams, err := mem.List()
	require.NoError(t, err)
	assert.Equal(t, store.INTERRUPTED, streams[0].State)

	// more than 1s because scheduler should have time to switch
	time.Sleep(1100 * time.Millisecond)

	streams, err = mem.List()
	require.NoError(t, err)
	assert.Equal(t, store.FINISHED, streams[0].State)
}

func TestInmemStore_Interrupt_NotFound(t *testing.T) {
	mem := prep(t)

	err := mem.Interrupt("8dff7c72-3edb-4718-87e9-6d60f653b4cf")
	assert.Error(t, err)
}

func TestInmemStore_Interrupt_Created(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	err = mem.Interrupt(stream.ID)
	assert.Error(t, err)
}

func TestInmemStore_Interrupt_Interrupted(t *testing.T) {
	mem := prep(t)

	stream, _ := mem.Create()
	mem.Start(stream.ID)

	err := mem.Interrupt(stream.ID)
	require.NoError(t, err)

	err = mem.Interrupt(stream.ID)
	assert.Error(t, err)
}

func TestInmemStore_Interrupt_Finished(t *testing.T) {
	mem := prep(t)

	stream, _ := mem.Create()
	mem.Start(stream.ID)

	err := mem.Interrupt(stream.ID)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = mem.Interrupt(stream.ID)
	assert.Error(t, err)
}

func TestInmemStore_StartInterruptStart(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	err = mem.Start(stream.ID)
	require.NoError(t, err)

	err = mem.Interrupt(stream.ID)
	require.NoError(t, err)

	// start again
	err = mem.Start(stream.ID)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	streams, err := mem.List()
	require.NoError(t, err)
	assert.Equal(t, store.ACTIVE, streams[0].State)
}

func TestInmemStore_StartInterruptAggressive(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mem.Start(stream.ID)
			mem.Interrupt(stream.ID)
		}()
	}
	wg.Wait()

	streams, err := mem.List()
	require.NoError(t, err)
	assert.Equal(t, store.INTERRUPTED, streams[0].State)
}

func TestInmemStore_StartInterruptConcurrent(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			mem.Start(stream.ID)
		}()

		go func() {
			defer wg.Done()
			mem.Interrupt(stream.ID)
		}()
	}
	wg.Wait()

	streams, err := mem.List()
	require.NoError(t, err)
	assert.True(t, streams[0].State == store.INTERRUPTED || streams[0].State == store.ACTIVE)
}

func TestInmemStore_Delete(t *testing.T) {
	mem := prep(t)

	stream, err := mem.Create()
	require.NoError(t, err)

	err = mem.Delete(stream.ID)
	assert.NoError(t, err)

	streams, err := mem.List()
	require.NoError(t, err)
	assert.Equal(t, 0, len(streams))
}

func TestInmemStore_Delete_NotFound(t *testing.T) {
	mem := prep(t)

	err := mem.Delete("8dff7c72-3edb-4718-87e9-6d60f653b4cf")
	assert.Error(t, err)
}

func prep(t *testing.T) store.Engine {
	return NewStore(1)
}

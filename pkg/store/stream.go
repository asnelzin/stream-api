package store

const (
	CREATED     = "created"
	ACTIVE      = "active"
	INTERRUPTED = "interrupted"
	FINISHED    = "finished"
)

type Stream struct {
	ID      string `jsonapi:"primary,stream"`
	State   string `jsonapi:"attr,state"`
	Created string `jsonapi:"attr,created"`
}

type Engine interface {
	Create() (*Stream, error)
	List() ([]*Stream, error)
	Start(id string) error
	Interrupt(id string) error
	Delete(id string) error
}

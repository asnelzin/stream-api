package rest

import (
	"net/http"

	"github.com/asnelzin/stream-api/pkg/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func MakeHandler(dataStore store.Engine) http.Handler {
	h := handler{dataStore}
	r := chi.NewRouter()

	r.Get("/", h.getAllStreams)
	r.Post("/", h.createStream)
	r.Post("/{streamID}/start", h.startStream)
	r.Post("/{streamID}/stop", h.stopStream)

	r.Delete("/{streamID}", h.deleteStream)
	return r
}

type handler struct {
	dataStore store.Engine
}

func (h handler) createStream(w http.ResponseWriter, r *http.Request) {
	stream, err := h.dataStore.Create()
	if err != nil {
		SendErrorJSONAPI(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	render.Status(r, http.StatusCreated)
	JSONAPI(w, r, stream)
}

func (h handler) getAllStreams(w http.ResponseWriter, r *http.Request) {
	streams, err := h.dataStore.List()
	if err != nil {
		SendErrorJSONAPI(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}

	JSONAPI(w, r, streams)
}

func (h handler) startStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "streamID")
	err := h.dataStore.Start(streamID)
	if err != nil {
		SendErrorJSONAPI(w, http.StatusBadRequest, "start_stream", err.Error())
		return
	}
	w.WriteHeader(200)
}

func (h handler) stopStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "streamID")

	err := h.dataStore.Interrupt(streamID)
	if err != nil {
		SendErrorJSONAPI(w, http.StatusBadRequest, "stop_stream", err.Error())
		return
	}
	w.WriteHeader(200)
}

func (h handler) deleteStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "streamID")

	err := h.dataStore.Delete(streamID)
	if err != nil {
		SendErrorJSONAPI(w, http.StatusBadRequest, "delete_stream", err.Error())
		return
	}
	w.WriteHeader(200)
}

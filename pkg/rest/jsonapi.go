package rest

import (
	"bytes"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/jsonapi"
	"strconv"
)

// JSONAPI marshals 'v' to JSON-API and setting the
// Content-Type as application/vnd.api+json.
func JSONAPI(w http.ResponseWriter, r *http.Request, v interface{}) {
	buf := &bytes.Buffer{}

	if err := jsonapi.MarshalPayload(buf, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.api+json; charset=utf-8")
	if status, ok := r.Context().Value(render.StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf.Bytes())
}

// SendErrorJSONAPI makes JSON-API body and responds with error code
func SendErrorJSONAPI(w http.ResponseWriter, code int, title string, details string) {
	// log.Printf("[DEBUG] %s", errDetailsMsg(r, code, err, details))
	w.Header().Set("Content-Type", "application/vnd.api+json; charset=utf-8")
	w.WriteHeader(code)
	jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
		Title:  title,
		Detail: details,
		Status: strconv.Itoa(code),
	}})
}

package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/asnelzin/stream-api/pkg/store"
	"github.com/asnelzin/stream-api/pkg/store/inmem"
	"github.com/google/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Ping(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	body, code := get(t, ts.URL+"/ping")
	assert.Equal(t, "pong", body)
	assert.Equal(t, 200, code)
}

func TestHandler_List(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	stream := createStream(t, ts)

	res, code := get(t, ts.URL+"/v1/streams/")
	assert.Equal(t, 200, code)

	streams, err := jsonapi.UnmarshalManyPayload(
		bytes.NewReader([]byte(res)), reflect.TypeOf(new(store.Stream)))
	require.Nil(t, err)
	assert.Equal(t, 1, len(streams))
	assert.Equal(t, stream, *streams[0].(*store.Stream))
}

func TestHandler_ListEmpty(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	res, code := get(t, ts.URL+"/v1/streams/")
	assert.Equal(t, 200, code)

	assert.Equal(t, "{\"data\":[]}\n", res)
}

func TestHandler_Create(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	resp, err := post(t, ts.URL+"/v1/streams/", "")
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var stream store.Stream
	err = jsonapi.UnmarshalPayload(resp.Body, &stream)
	assert.Nil(t, err)
	assert.Equal(t, stream.State, "created")
}

func TestHandler_Start(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	stream := createStream(t, ts)

	resp, err := post(t, ts.URL+"/v1/streams/"+stream.ID+"/start", "")
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestHandler_Start_Error(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	resp, err := post(t, ts.URL+"/v1/streams/8dff7c72-3edb-4718-87e9-6d60f653b4cf/start", "")
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "{\"errors\":[{\"title\":\"start_stream\","+
		"\"detail\":\"could not find stream with id 8dff7c72-3edb-4718-87e9-6d60f653b4cf\","+
		"\"status\":\"400\"}]}\n", string(body))
}

func TestHandler_Stop(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	stream := createStream(t, ts)
	resp, err := post(t, ts.URL+"/v1/streams/"+stream.ID+"/start", "")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	resp, err = post(t, ts.URL+"/v1/streams/"+stream.ID+"/stop", "")
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestHandler_Stop_Error(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	resp, err := post(t, ts.URL+"/v1/streams/8dff7c72-3edb-4718-87e9-6d60f653b4cf/stop", "")
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "{\"errors\":[{\"title\":\"stop_stream\","+
		"\"detail\":\"could not find stream with id 8dff7c72-3edb-4718-87e9-6d60f653b4cf\","+
		"\"status\":\"400\"}]}\n", string(body))
}

func TestHandler_Delete(t *testing.T) {
	srv, ts := prep(t)
	require.NotNil(t, srv)
	defer cleanup(ts)

	stream := createStream(t, ts)

	r, err := http.NewRequest("DELETE", ts.URL+"/v1/streams/"+stream.ID, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "", string(body))

	res, _ := get(t, ts.URL+"/v1/streams/")
	assert.Equal(t, "{\"data\":[]}\n", res)
}

func createStream(t *testing.T, ts *httptest.Server) store.Stream {
	resp, err := post(t, ts.URL+"/v1/streams/", "")
	require.Nil(t, err)

	var stream store.Stream
	err = jsonapi.UnmarshalPayload(resp.Body, &stream)
	require.Nil(t, err)

	return stream
}

func get(t *testing.T, url string) (string, int) {
	r, err := http.Get(url)
	require.Nil(t, err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	require.Nil(t, err)
	return string(body), r.StatusCode
}

func post(t *testing.T, url string, body string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	assert.Nil(t, err)
	req.SetBasicAuth("dev", "password")
	return client.Do(req)
}

func prep(t *testing.T) (*Server, *httptest.Server) {
	dataStore := inmem.NewStore(3)
	srv := &Server{
		DataStore: dataStore,
	}
	ts := httptest.NewServer(srv.routes())
	return srv, ts
}

func cleanup(ts *httptest.Server) {
	ts.Close()
}

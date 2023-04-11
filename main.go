package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/rs/cors"
)

var (
	ccTimeline          Cache
	ccPhoto             Cache
	errNoIDProvided     = errors.New("no ID provided")
	internalServerError = http.StatusInternalServerError
)

type Dot struct {
	Title   string  `json:"title"`
	Descr   string  `json:"descr,omitempty"`
	PhotoID string  `json:"photo_id,omitempty"`
	Size    float64 `json:"size"`
	Epoch   int64   `json:"epoch"`
}

type Timeline struct {
	Alias string         `json:"alias,omitempty"`
	ID    string         `json:"timelineID"`
	Dots  map[string]Dot `json:"dots,omitempty"`
}

type User struct {
	ID        string   `json:"id"`
	Timelines []string `json:"timelines"`
}

type Base struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type GetResponse struct {
	Base
	TimelineContent Timeline `json:"timeline_content,omitempty"`
}

func response(w http.ResponseWriter, err error) {
	var res Base

	if err != nil {
		w.WriteHeader(internalServerError)
		res = Base{Ok: false, Error: err.Error()}
	} else {
		res = Base{Ok: true}
	}

	jsn, jsnErr := json.Marshal(res)
	if jsnErr != nil {
		log.Println("response", "json.Marshal", err)
		return
	}

	if _, wrErr := w.Write(jsn); wrErr != nil {
		log.Println("response", "w.Write", err)
	}
}

func encode(tl Timeline) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(tl); err != nil {
		return nil, fmt.Errorf("enc.Encode: %w", err)
	}
	return buf.Bytes(), nil
}

func decode(b []byte) (Timeline, error) {
	var buf bytes.Buffer
	var tl Timeline

	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&tl); err != nil {
		return Timeline{}, fmt.Errorf("dec.Decode: %w", err)
	}
	return tl, nil
}

// writes Base: ok if Timeline is successfully inserted in database
func putHandler(rw http.ResponseWriter, rq *http.Request) {
	var (
		tl Timeline
		id string
	)

	if id = rq.URL.Query().Get("id"); id == "" {
		log.Println("putHandler", "rq.URL.Query().Get", errNoIDProvided)
		response(rw, errNoIDProvided)
		return
	}

	body, err := io.ReadAll(rq.Body)
	if err != nil {
		log.Println("putHandler", "io.ReadAll", err)
		response(rw, err)
		return
	}

	if err := json.Unmarshal(body, &tl); err != nil {
		log.Println("putHandler", "json.Unmarshal", err)
		response(rw, err)
		return
	}

	b, err := encode(tl)
	if err != nil {
		log.Println("putHandler", "encode", err)
		response(rw, err)
		return
	}

	if err := ccTimeline.Put(id, string(b)); err != nil {
		log.Println("putHandler", "ccTimeline.Put", err)
		response(rw, err)
		return
	}

	response(rw, nil)
}

func getHandler(rw http.ResponseWriter, rq *http.Request) {
	var (
		tl Timeline
		id string
	)

	if id = rq.URL.Query().Get("timeline-ID"); id == "" {
		log.Println("getHandler", "rq.URL.Query().Get", errNoIDProvided)
		response(rw, errNoIDProvided)
		return
	}

	b, err := ccTimeline.Get(id)
	if err != nil {
		log.Println("getHandler", "ccTimeline.Get", err)
		response(rw, err)
		return
	}

	if tl, err = decode(b); err != nil {
		log.Println("getHandler", "decode", err)
		response(rw, err)
		return
	}

	jsn, err := json.Marshal(GetResponse{
		Base:            Base{Ok: true},
		TimelineContent: tl,
	})
	if err != nil {
		log.Println("getHandler", "decode", err)
		response(rw, err)
		return
	}

	if _, err = rw.Write(jsn); err != nil {
		log.Println("getHandler", "rw.Write", err)
		response(rw, err)
		return
	}
}

// if a timeline-ID is specified but not a dot-ID, it means that the whole
// timeline should be deleted, if both are specified, than only the single dot
// inside the timeline should be deleted, and a new timeline without that dot
// should be returned. If a timeline-ID is not specified, than del returns errNoIDProvided
func delHandler(rw http.ResponseWriter, rq *http.Request) {
	tlID := rq.URL.Query().Get("id")
	if tlID == "" {
		log.Println("delHandler", "rq.URL.Query().Get", errNoIDProvided)
		response(rw, errNoIDProvided)
		return
	}

	dotID := rq.URL.Query().Get("dot-id")
	err := ccTimeline.Del(tlID, dotID)
	if err != nil {
		log.Println("delHandler", "ccTimeline.Del", err)
		response(rw, err)
		return
	}
	response(rw, nil)
}

func main() {
	var (
		mux  = http.NewServeMux()
		port = "8080"
	)

	ucd, err := os.UserCacheDir()
	if err != nil {
		log.Fatalln("main", "os.UserCacheDir", err)
	}

	ccTimeline = Cache(path.Join(ucd, "timeline-for-humanity", "timelines"))
	ccPhoto = Cache(path.Join(ucd, "timeline-for-humanity", "photos"))

	mux.HandleFunc("/get", getHandler)
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/del", delHandler)

	log.Printf("running on port %s...", port)
	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler))
}

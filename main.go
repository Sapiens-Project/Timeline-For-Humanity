package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/rs/cors"
)

var (
	ccTimeline      Cache
	ccPhoto         Cache
	errNoIDProvided = errors.New("no ID provided")
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
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type GetResponse struct {
	Base
	Timeline Timeline `json:"timeline_content,omitempty"`
}

func okres(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	jsn, err := json.Marshal(Base{OK: true})
	if err != nil {
		log.Println("okres", "json.Marshal", err)
		return
	}
	if _, err := w.Write(jsn); err != nil {
		log.Println("okres", "w.Write", err)
	}
}

func eres(w http.ResponseWriter, err error) {
	if err == nil {
		okres(w)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	jsn, err := json.Marshal(Base{OK: false, Error: err.Error()})
	if err != nil {
		log.Println("eres", "json.Marshal", err)
		return
	}
	if _, err := w.Write(jsn); err != nil {
		log.Println("eres", "w.Write", err)
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
	var (
		tl  Timeline
		buf bytes.Buffer
	)

	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&tl); err != nil {
		return Timeline{}, fmt.Errorf("dec.Decode: %w", err)
	}
	return tl, nil
}

func putTimeline(id string, tl Timeline) error {
	b, err := encode(tl)
	if err != nil {
		return fmt.Errorf("saveTimeline: %w", err)
	}
	return ccTimeline.Put([]byte(id), b)
}

func getTimeline(id string) (Timeline, error) {
	b, err := ccTimeline.Get([]byte(id))
	if err != nil {
		return Timeline{}, fmt.Errorf("getTimeline: %w", err)
	}
	return decode(b)
}

func delTimeline(id string) error {
	return ccTimeline.Del([]byte(id))
}

// putHandler updates an existing timeline with the one provided, it also
// creates the timeline if it didn't exist.
func putHandler(w http.ResponseWriter, rq *http.Request) {
	var tl Timeline

	id := rq.URL.Query().Get("id")
	if id == "" {
		log.Println("putHandler", "rq.URL.Query().Get", errNoIDProvided)
		eres(w, errNoIDProvided)
		return
	}

	body, err := io.ReadAll(rq.Body)
	if err != nil {
		log.Println("putHandler", "io.ReadAll", err)
		eres(w, err)
		return
	}

	if err := json.Unmarshal(body, &tl); err != nil {
		log.Println("putHandler", "json.Unmarshal", err)
		eres(w, err)
		return
	}

	if err := putTimeline(id, tl); err != nil {
		log.Println("putHandler", "putTimeline", err)
		eres(w, err)
		return
	}
	okres(w)
}

// getHandler responds with the timeline associated with the given ID.
func getHandler(w http.ResponseWriter, rq *http.Request) {
	id := rq.URL.Query().Get("id")
	if id == "" {
		log.Println("getHandler", "rq.URL.Query().Get", errNoIDProvided)
		eres(w, errNoIDProvided)
		return
	}

	timeline, err := getTimeline(id)
	if err != nil {
		log.Println("getHandler", "getTimeline", err)
		eres(w, err)
		return
	}

	jsn, err := json.Marshal(GetResponse{
		Base:     Base{OK: true},
		Timeline: timeline,
	})
	if err != nil {
		log.Println("getHandler", "decode", err)
		eres(w, err)
		return
	}

	if _, err = w.Write(jsn); err != nil {
		log.Println("getHandler", "w.Write", err)
		return
	}
}

// delHandler deletes a timeline from the timeline cache.
func delHandler(w http.ResponseWriter, rq *http.Request) {
	id := rq.URL.Query().Get("id")
	if id == "" {
		log.Println("delHandler", "rq.URL.Query().Get", errNoIDProvided)
		eres(w, errNoIDProvided)
		return
	}
	if err := delTimeline(id); err != nil {
		log.Println("delHandler", "delTimeline", err)
		eres(w, err)
		return
	}
	okres(w)
}

func main() {
	var (
		port string
		mux  = http.NewServeMux()
	)

	flag.StringVar(&port, "p", ":8080", "The port the server will be listening to.")
	flag.Parse()

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
	log.Fatal(http.ListenAndServe(port, handler))
}

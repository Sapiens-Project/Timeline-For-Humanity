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
	Alias string `json:"alias,omitempty"`
	//User  string         `json:"user"`
	ID   string         `json:"timelineID"`
	Dots map[string]Dot `json:"dots"`
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

func errorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	res := Base{Ok: false, Error: err.Error()}
	jsn, err := json.Marshal(res)
	if err != nil {
		log.Println("errorResponse", "json.Marshal", err)
		return
	}

	if _, err := w.Write(jsn); err != nil {
		log.Println("errorResponse", "w.Write", err)
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

// da "esportare" un po' di roba dalla funzione
// la chiave da dove la prende?
// rq.Body è un ioReader quindi non si può usare direttamente, la funzione
// io.ReadAll prende il contenuto di un ioReader e lo mettere in una var []byte
// così si può usare per ottenere il json
func putHandler(rw http.ResponseWriter, rq *http.Request) {
	var (
		tl Timeline
		id string
	)

	if id = rq.URL.Query().Get("timeline-ID"); id == "" {
		log.Println("putHandler", "rq.URL.Query().Get", errNoIDProvided)
		errorResponse(rw, errNoIDProvided)
		return
	}

	// per ogni errore deve mandare una risposta negativa al frontend
	body, err := io.ReadAll(rq.Body)
	if err != nil {
		log.Println("putHandler", "io.ReadAll", err)
		errorResponse(rw, err)
		return
	}

	if err := json.Unmarshal(body, &tl); err != nil {
		log.Println("putHandler", "json.Unmarshal", err)
		errorResponse(rw, err)
		return
	}

	b, err := encode(tl)
	if err != nil {
		log.Println("putHandler", "encode", err)
		errorResponse(rw, err)
		return
	}

	if err := ccTimeline.Put([]byte(id), b); err != nil {
		log.Println("putHandler", "ccTimeline.Put", err)
		errorResponse(rw, err)
		return
	}
}

func getHandler(rw http.ResponseWriter, rq *http.Request) {
	// richiede la timeline con UUID
	// controlla UUID nel db
	// restituisce un json con le info dal db
	errorResponse(rw, errors.New("dio porco"))
}

func delHandler(rw http.ResponseWriter, rq *http.Request) {

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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	firebase "firebase.google.com/go"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type StatsWriter struct {
	StartTime    time.Time         `json:"startTime"`
	RequestCount int               `json:"requestCount"`
	Server       *http.ServeMux    `json:"-"`
	Logger       *log.Logger       `json:"-"`
	FSClient     *firestore.Client `json:"-"`
	Filename     string            `json:"filename"`
	Size         int               `json:"size"`
	BytesWritten int               `json:"bytes"`
	ID           string            `json:"id"`
	Stats        []Stat            `json:"stats"`
	Mux          *sync.RWMutex     `json:"-"`
}

type Stat struct {
	Value []float64     `json:"value"`
	Time  time.Time     `json:"time"`
	ID    string        `json:"id"`
	Extra []interface{} `json:"extra"`
}

func NewStatsWriter(noDocker bool, filename string, firestoreMode bool) (*StatsWriter, error) {
	var fname string
	var client *firestore.Client
	if noDocker {
		fname = filename
	} else {
		fname = fmt.Sprintf("/logs/%v", filename)
	}
	if firestoreMode {
		ctx := context.Background()
		sa := option.WithCredentialsFile("/fbase.json") // Path to service account key
		app, err := firebase.NewApp(ctx, nil, sa)
		if err != nil {
			// Handle error
		}

		client, err := app.Firestore(ctx)
		if err != nil {
			// Handle error
		}
		defer client.Close()

	}
	fh, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	newLog := log.New(fh, "", log.LstdFlags)
	newLog.SetFlags(0)
	lock := &sync.RWMutex{}
	stats := make([]Stat, 0)
	return &StatsWriter{
		StartTime: time.Now(),
		Logger:    newLog,
		Mux:       lock,
		Stats:     stats,
		Filename:  filename,
		FSClient:  client,
		ID:        RandomUUID(),
	}, nil
}

func (s *StatsWriter) RootHandler(w http.ResponseWriter, r *http.Request) {

	// log.Println("root handler", s.ID)
	var in Stat
	err := ReadJSON(w, r, &in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// s.AppendStats(in)

	out, err := json.Marshal(in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if s.FSClient != nil {
		_, _, err := s.FSClient.Collection(s.Filename).Add(context.Background(), in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		s.Logger.Println(string(out))
	}
	s.RequestCount++
	fmt.Fprintf(w, "OK\n")
}

func (s *StatsWriter) SaveToFireStore(ctx context.Context, database string) error {
	return nil
}

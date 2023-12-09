package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type StatsWriter struct {
	StartTime    time.Time      `json:"startTime"`
	RequestCount int            `json:"requestCount"`
	Server       *http.ServeMux `json:"-"`
	Logger       *log.Logger    `json:"-"`
	Filename     string         `json:"filename"`
	Size         int            `json:"size"`
	BytesWritten int            `json:"bytes"`
	ID           string         `json:"id"`
	Stats        []Stat         `json:"stats"`
	Mux          *sync.RWMutex  `json:"-"`
}

type Stat struct {
	Value float64   `json:"value"`
	Time  time.Time `json:"time"`
	ID    string    `json:"id"`
}

func NewStatsWriter(filename string) (*StatsWriter, error) {
	fh, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
		ID:        RandomUUID(),
	}, nil
}

func (s *StatsWriter) RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("root handler", s.ID)
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
	s.Logger.Println(string(out))
	s.RequestCount++
	fmt.Fprintf(w, "OK\n")
}

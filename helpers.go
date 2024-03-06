package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func RandomUUID() string {
	res := uuid.New()
	return res.String()
}

func ReadJSON(w http.ResponseWriter, r *http.Request, out interface{}) error {
	maxBytes := 4024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&out)
	if err != nil {
		return err
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return err
	}
	return nil
}

func (m *MainServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimLeft(r.URL.Path, "/")
	bits := strings.Split(path, "/")
	if len(bits) < 1 {
		m.Logger.Println("invalid path", path)
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	m.Mux.Lock()
	svr, ok := m.Writers[bits[0]]
	m.Visits++
	if !ok {
		m.Logger.Println("creating new stats writer", bits[0])
		svr, err := NewStatsWriter(m.Modes, fmt.Sprintf("%s.log", bits[0]), m.FSClient, DetectMode(m.Modes))
		if err != nil {
			m.Logger.Println(err, "error creating stats writer")
			http.Error(w, "error creating stats writer", http.StatusInternalServerError)
			return
		}
		m.Writers[bits[0]] = svr
		m.Mux.Unlock()
		svr.RootHandler(w, r)

	} else {
		m.Mux.Unlock()
		svr.RootHandler(w, r)
	}
	// m.Logger.Printf("served request (%v) for %v", path, r.RemoteAddr)
}

func (s *StatsWriter) AppendStats(stat Stat) {
	// only append the list if the lock is acquired
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.Stats = append(s.Stats, stat)
}

func (s *StatsWriter) WriteStatsToLogger() {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for _, stat := range s.Stats {
		out, err := json.Marshal(stat)
		if err != nil {
			fmt.Println(err, "error marshalling stat", s.ID)
			continue
		}
		s.Logger.Println(string(out))
	}
	s.Stats = []Stat{}
}

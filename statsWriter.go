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

	"cloud.google.com/go/firestore"
)

type StatsWriter struct {
	HybridWriter
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

type HybridWriter interface {
	Write(stat ...Stat) (int, error)
	Setup(fname string, ifc interface{}) error
}

type FirestoreWriter struct {
	Filename string
	FSClient *firestore.Client
}

func (f *FirestoreWriter) Write(stat ...Stat) (int, error) {
	ctx := context.Background()
	for _, s := range stat {
		_, _, err := f.FSClient.Collection(f.Filename).Add(ctx, s)
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func (f *FirestoreWriter) Setup(fname string, ifc interface{}) error {
	fmt.Println("running setup in firestore mode")
	switch ifc.(type) {
	case *firestore.Client:
		f.FSClient = ifc.(*firestore.Client)
		f.Filename = fname
		return nil
	}

	return fmt.Errorf("invalid type %T", ifc)
}

type InMemoryWriter struct {
	Filename       string
	Bucket         string
	WriteFrequency time.Duration
}

type inMemoryWriterConfig struct {
	Filename       string
	Bucket         string
	WriteFrequency time.Duration
}

func (i *InMemoryWriter) Write(stat ...Stat) (int, error) {
	return 0, nil
}

func (i *InMemoryWriter) Setup(fname string, ifc interface{}) error {
	fmt.Println("running setup in in-memory mode")
	switch ifc.(type) {
	case inMemoryWriterConfig:
		cfg := ifc.(inMemoryWriterConfig)
		i.Filename = cfg.Filename
		i.Bucket = cfg.Bucket
		i.WriteFrequency = cfg.WriteFrequency
		return nil
	}
	return fmt.Errorf("invalid type %T", ifc)
}

type SimpleWriter struct {
	Filename string
	Logger   *log.Logger
}

func (s *SimpleWriter) Write(stat ...Stat) (int, error) {
	for _, st := range stat {
		out, err := json.Marshal(st)
		if err != nil {
			return 0, err
		}
		s.Logger.Println(string(out))
	}
	return 0, nil
}

func (s *SimpleWriter) Setup(fname string, ifc interface{}) error {
	fmt.Println("running setup in simple mode")
	s.Filename = fname
	fh, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	newLog := log.New(fh, "", log.LstdFlags)
	newLog.SetFlags(0)
	s.Logger = newLog
	return nil
}

func NewStatsWriter(modes Modes, filename string, client *firestore.Client, hw HybridWriter) (*StatsWriter, error) {
	var fname string
	if modes.NoDocker {
		fname = filename
	} else {
		fname = fmt.Sprintf("/logs/%v", filename)
	}

	switch {
	case modes.FirestoreMode:
		fmt.Println("setting up firestore writer")
		err := hw.Setup(fname, client)
		if err != nil {
			return nil, err
		}
	case modes.InMemoryMode:
		err := hw.Setup(fname, inMemoryWriterConfig{
			Filename:       fname,
			Bucket:         "stats",
			WriteFrequency: 5 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	default:
		err := hw.Setup(fname, nil)
		if err != nil {
			return nil, err
		}
	}
	lock := &sync.RWMutex{}
	stats := make([]Stat, 0)
	return &StatsWriter{
		HybridWriter: hw,
		StartTime:    time.Now(),
		Mux:          lock,
		Stats:        stats,
		Filename:     filename,
		FSClient:     client,
		ID:           RandomUUID(),
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

	// out, err := json.Marshal(in)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	s.Write(in)
	s.RequestCount++
	fmt.Fprintf(w, "OK\n")
}

func (s *StatsWriter) SaveToFireStore(ctx context.Context, database string) error {
	return nil
}

func DetectMode(modes Modes) HybridWriter {
	switch {
	case modes.FirestoreMode:
		fmt.Println("activating firestore mode")
		return &FirestoreWriter{}
	case modes.InMemoryMode:
		fmt.Println("activating in memory mode")
		return &InMemoryWriter{}
	default:
		fmt.Println("activating simple mode")
		return &SimpleWriter{}
	}
}

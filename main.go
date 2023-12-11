package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	port = flag.Int("port", 20080, "port to listen on")
)

type MainServer struct {
	Logger    *log.Logger
	Port      int
	StartTime time.Time
	Visits    int
	Server    *http.Server
	Mux       *sync.RWMutex
	Writers   map[string]*StatsWriter
}

func main() {
	mx := &sync.RWMutex{}
	flag.Parse()
	app := &MainServer{
		Logger: log.New(os.Stdout, "main_server _ ", log.LstdFlags),
		Port:   *port,
		Mux:    mx,
	}
	app.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Port),
		Handler: app,
	}
	app.Writers = make(map[string]*StatsWriter)
	app.Logger.Printf("Starting server on port %d\n", app.Port)
	app.StartTime = time.Now()
	go func() {
		for range time.Tick(5 * time.Second) {
			if app.Visits > 0 {
				app.Mux.Lock()
				app.Logger.Printf("got %v visits in the last 5 seconds\r", app.Visits)
				app.Visits = 0
				app.Mux.Unlock()
			}
		}
	}()
	app.Logger.Fatal(app.Server.ListenAndServe())
}

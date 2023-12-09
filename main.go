package main

/*

88                                   88
88                                   ""
88
88 ,adPPYYba, 8b,dPPYba,   ,adPPYba, 88 88       88 88,dPYba,,adPYba,
88 ""     `Y8 88P'   `"8a a8"     "" 88 88       88 88P'   "88"    "8a
88 ,adPPPPP88 88       88 8b         88 88       88 88      88      88
88 88,    ,88 88       88 "8a,   ,aa 88 "8a,   ,a88 88      88      88
88 `"8bbdP"Y8 88       88  `"Ybbd8"' 88  `"YbbdP'Y8 88      88      88
_______________________________________________________________________
This is the main entry point of the app.
*/

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	port = flag.Int("port", 20080, "port to listen on")
)

type MainServer struct {
	Logger  *log.Logger
	Port    int
	Server  *http.Server
	Mux     *sync.RWMutex
	Writers map[string]*StatsWriter
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
	app.Logger.Fatal(app.Server.ListenAndServe())
}

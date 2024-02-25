package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
	port          = flag.Int("port", 20080, "port to listen on")
	noDocker      = flag.Bool("no-docker", false, "are we running in a container")
	firestoreMode = flag.Bool("firestore", true, "use firestore")
	projectId     = flag.String("project", "plated-dryad-148318", "project id")
)

type MainServer struct {
	Modes
	FSClient  *firestore.Client
	Logger    *log.Logger
	Port      int
	StartTime time.Time
	Visits    int
	Server    *http.Server
	Mux       *sync.RWMutex
	Writers   map[string]*StatsWriter
}

type Modes struct {
	NoDocker      bool
	FirestoreMode bool
}

func main() {
	mx := &sync.RWMutex{}
	client := &firestore.Client{}
	defer client.Close()
	flag.Parse()
	if *firestoreMode {
		ctx := context.Background()
		sa := option.WithCredentialsFile("/fbase.json")
		cfg := &firebase.Config{ProjectID: *projectId}
		fb, err := firebase.NewApp(ctx, cfg, sa)
		if err != nil {
			fmt.Println("error initializing app:", err)
			os.Exit(1)
		}

		client, err = fb.Firestore(ctx)
		if err != nil {
			fmt.Println("error getting Firestore client:", err)
			os.Exit(1)
		}
		fmt.Println("Firestore connected")

	}
	app := &MainServer{
		FSClient: client,
		Logger:   log.New(os.Stdout, "main_server _ ", log.LstdFlags),
		Port:     *port,
		Mux:      mx,
		Modes: Modes{
			NoDocker:      *noDocker,
			FirestoreMode: *firestoreMode,
		},
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

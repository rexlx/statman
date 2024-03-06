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
	firestoreMode = flag.Bool("firestore", false, "use firestore")
	projectId     = flag.String("project", "tubular-monkey-514321", "project id")
	inMemoryMode  = flag.Bool("in-memory", false, "use in-memory storage")
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
	InMemoryMode  bool
}

func main() {
	flag.Parse()
	mx := &sync.RWMutex{}

	client := &firestore.Client{}
	defer client.Close()

	if *firestoreMode {
		ctx := context.Background()
		cfg := &firebase.Config{ProjectID: *projectId}
		sa := option.WithCredentialsFile("/Users/rxlx/bin/data/fbase.json")
		fb, err := firebase.NewApp(ctx, cfg, sa)
		// fb, err := firebase.NewApp(ctx, nil)
		if err != nil {
			fmt.Println("error initializing app in firestore mode:", err)
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
			InMemoryMode:  *inMemoryMode,
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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/runtime/middleware"
	httpServer "github.com/umgbhalla/gokv/api/http"
	wsServer "github.com/umgbhalla/gokv/api/websocket"
	"github.com/umgbhalla/gokv/internal/persistence"
	"github.com/umgbhalla/gokv/internal/query"
	"github.com/umgbhalla/gokv/internal/store"
)

// @title GoKV API
// @version 1.0
// @description This is the API server for the GoKV key-value store.
// @host localhost:8080
// @BasePath /

func main() {
	logFile, err := os.OpenFile("gokv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Starting GoKV server...")

	kvStore := store.New()
	kvQuery := query.New(kvStore)

	persister := persistence.New(kvStore, "data.json", 30*time.Second)
	if err := persister.Load(); err != nil {
		log.Printf("Error loading data: %v", err)
	}
	go persister.Start()

	httpSrv := httpServer.NewServer(kvStore, kvQuery)

	wsSrv := wsServer.NewServer(kvStore, kvQuery)

	opts := middleware.SwaggerUIOpts{SpecURL: "/swagger.json"}
	sh := middleware.SwaggerUI(opts, nil)
	httpSrv.Router().Handle("/docs", sh)
	httpSrv.Router().Handle("/swagger.json", http.FileServer(http.Dir("./docs")))

	go func() {
		log.Printf("Starting HTTP server on :8080")
		if err := httpSrv.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting WebSocket server on :8081")
		if err := wsSrv.Start(":8081"); err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := wsSrv.Shutdown(ctx); err != nil {
		log.Printf("WebSocket server shutdown error: %v", err)
	}

	persister.Stop()

	if err := persister.Save(); err != nil {
		log.Printf("Error saving final state: %v", err)
	}

	log.Println("Shutdown complete")
}

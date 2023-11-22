package main

import (
	"dev-payment-gate/api/router"
	"dev-payment-gate/internal/app"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Initialize the application
    if err := app.Initialize("./"); err != nil {
        log.Fatalf("[Error] %v", err)
    }

	// Initialize the server
	server := http.Server{
		Addr:	 fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: router.Router(),
	}

	// Create a channel to receive interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	
	// Create a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Goroutine to handle interrupt signal
	go func() {
		// Wait for the interrupt signal
		<-interrupt

		// Increment WaitGroup counter to indicate the start of this goroutine
		wg.Add(1)

		// Clean up resources and gracefully exit
		if err := app.Clean(&server); err != nil {
			log.Printf("[Warning] %v", err)
		}

		// Exit the program
		os.Exit(0)
	}()

	// Start the server
	log.Printf("Listening to port %s for HTTP requests...\n", os.Getenv("PORT"))
	if err := server.ListenAndServe(); err != nil {
		log.Printf("%v", err)
	}

	// Wait for priority and exit
	wg.Wait()
	if err := app.Clean(&server); err != nil {
		log.Printf("[Warning] %v", err)
	}
	os.Exit(1)
}

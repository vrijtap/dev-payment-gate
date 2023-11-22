package app

import (
	"context"
	"dev-payment-gate/utils/database"
	"dev-payment-gate/web/templates"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Initialize initializes the application
func Initialize(relativeRootFolder string) error {
    // Load configurations from .env file
    if err := godotenv.Load(fmt.Sprintf("%s.env", relativeRootFolder)); err != nil {
        return fmt.Errorf("failed to load environment configurations from .env file: %v", err)
    }

    // Load templates from the templates folder
    if err := templates.Load(fmt.Sprintf("%sweb/templates/", relativeRootFolder)); err != nil {
        return fmt.Errorf("failed to load .html templates: %v", err)
    }

    // Initialize the database connection
    if err := database.Connect(os.Getenv("MONGO_URI"), "dev-payment-gate"); err != nil {
        return fmt.Errorf("unable to establish connection to the database: %v", err)
    }

    return nil
}

// Clean is a function that performs cleanup operations, closing the server and disconnecting from the database.
func Clean(server *http.Server) error {
    log.Println("Shutting down gracefully...")
    var errs []error

    // Attempt to close the HTTP server
    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        errs = append(errs, fmt.Errorf("unable to shutdown the server: %v", err))
    }

    // Attempt to disconnect from the database
    if err := database.Disconnect(); err != nil {
        errs = append(errs, fmt.Errorf("unable to disconnect the database: %v", err))
    }

	// No errors, return nil
    if len(errs) == 0 {
        return nil
    }

    // If there are errors, return them as a multi-error
    return errs[0]
}

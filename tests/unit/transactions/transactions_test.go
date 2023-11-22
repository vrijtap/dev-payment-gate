package transactions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"dev-payment-gate/api/router"
	"dev-payment-gate/internal/app"
	"dev-payment-gate/utils/model/transactions"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	r *mux.Router
	transactionInput transactions.TransactionInput
	transactionID primitive.ObjectID
)

// createResponse holds the response for TestCreate
type redirectResponse struct {
	URL string `json:"url"`
}

// TestMain functions as the entry point for our unit tests
func TestMain(m *testing.M) {
    // Setup the test server before running tests
    if err := app.Initialize("../../../"); err != nil {
        log.Fatalf("Initialization error: %v", err)
    }

    // Initialize the transaction router
    r = router.Router()

    // Run the tests
    exitCode := m.Run()

    // Exit with the status code from tests
    os.Exit(exitCode)
}

// TestNotFound tests behavior for invalid routes
func TestNotFound(t *testing.T) {
    // Create a request with a specific URI
	request := httptest.NewRequest("GET", "/nonexistent", nil)

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(recorder, request)

	// Validate the response
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("Expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

// TestCreate tests transaction creation for our API
func TestCreate(t *testing.T) {
	transactionInput = transactions.TransactionInput{
		Amount: 4.95,
		WebhookURL: "url",
		WebhookKey: "key",
		RedirectURL: "https://test.nl",
	}

	// Marshal transactionInput into JSON
	body, err := json.Marshal(transactionInput)
	if err != nil {
		t.Errorf("Error marshaling transactionInput: %v", err)
		return
	}

	// Create a request with a specific URI
	request := httptest.NewRequest("POST", "/transaction", bytes.NewBuffer(body))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")))
	request.Header.Add("Content-Type", "application/json")

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(recorder, request)

	// Validate the response
	if recorder.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, recorder.Code)
	}

	// Parse the JSON response into the response struct
	var redirection redirectResponse
	if err := json.NewDecoder(recorder.Body).Decode(&redirection); err != nil {
		t.Errorf("Error parsing JSON response: %v", err)
		return
	}

	// Parse the JSON response URL
	pathSegments := strings.Split(redirection.URL, "/")
	if len(pathSegments) != 5 {
		t.Fatalf("Invalid URL format: %s", redirection.URL)
	}
	id := pathSegments[len(pathSegments)-1]

	// Parse the ID into an ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		t.Fatalf("Invalid ObjectID format: %s", id)
	}
	transactionID = objectID
}

// TestTransactionDatabaseEntry checks if the transaction was saved correctly
func TestTransactionDatabaseEntry(t *testing.T) {
	// Fetch the transaction from the database
	transaction, err := transactions.GetByID(context.TODO(), transactionID)
	if err != nil {
		t.Fatalf("Could not fetch transaction from database: %v", err)
	}

	// Compare transaction fields with the expected input
	if transaction.ID != transactionID {
		t.Fatalf("Expected ID: %s, Got: %s", transactionID, transaction.ID)
	}
	if transaction.Amount != transactionInput.Amount {
		t.Errorf("Expected Amount: %f, Got: %f", transactionInput.Amount, transaction.Amount)
	}
	if transaction.WebhookURL != transactionInput.WebhookURL {
		t.Errorf("Expected WebhookURL: %s, Got: %s", transactionInput.WebhookURL, transaction.WebhookURL)
	}
	if transaction.WebhookKey != transactionInput.WebhookKey {
		t.Errorf("Expected WebhookKey: %s, Got: %s", transactionInput.WebhookKey, transaction.WebhookKey)
	}
	if transaction.RedirectURL != transactionInput.RedirectURL {
		t.Errorf("Expected RedirectURL: %s, Got: %s", transactionInput.RedirectURL, transaction.RedirectURL)
	}
}

// TestGetHTML verifies that the API can serve our HTML template
func TestGetHTML(t *testing.T) {
    // Create a request with a specific URI
	request := httptest.NewRequest("GET", fmt.Sprintf("/transaction/%s", transactionID.Hex()), nil)

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(recorder, request)

    // Validate the response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, recorder.Code)
	}

	// Define the required security headers
	requiredHeaders := []string{
		"Content-Security-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Feature-Policy",
	}

	// Verify the presence of security headers
	for _, header := range requiredHeaders {
		if value := recorder.Header().Get(header); value == "" {
			t.Errorf("Missing required security header: %s", header)
		}
	}

	// Verify that the response content type is HTML
	contentType := recorder.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected HTML content, but got Content-Type: %s", contentType)
	}
}

// TestGetJS verifies that the APi can serve our JS template
func TestGetJS(t *testing.T) {
	// Create a request with a specific URI
	request := httptest.NewRequest("GET", fmt.Sprintf("/transaction/js/%s", transactionID.Hex()), nil)

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(recorder, request)

    // Validate the response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, recorder.Code)
	}

	// Verify that the response content type is JS
	contentType := recorder.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/javascript") {
		t.Errorf("Expected JS content, but got Content-Type: %s", contentType)
	}
}



// TestProcess tests the processing function for fake payments and the redirect url
func TestProcess(t *testing.T) {
	// Create a request with a specific URI
	request := httptest.NewRequest("POST", fmt.Sprintf("/transaction/%s", transactionID.Hex()), nil)

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(recorder, request)

	// Validate the response
	if recorder.Code != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, recorder.Code)
	}

	// Parse the JSON response into the response struct
	var redirection redirectResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &redirection); err != nil {
		t.Errorf("Error parsing JSON response: %v", err)
	}

	// Verify that the redirection URL is correct
	if redirection.URL != transactionInput.RedirectURL {
		t.Errorf("Expected %s, but got: %s", transactionInput.RedirectURL, redirection.URL)
	}
}

// TestDeleted checks if the transaction was deleted correctly
func TestDeleted(t *testing.T) {
	// Fetch the transaction from the database to test if it still exists
	_, err := transactions.GetByID(context.TODO(), transactionID)
	if err == nil {
		t.Error("Successfully fetched transaction after deletion")
	}
}

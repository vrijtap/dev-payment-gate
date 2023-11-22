package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"dev-payment-gate/utils/model/transactions"
	"dev-payment-gate/web/templates"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StatusData holds the data for a transaction
type StatusData struct {
    Status string `json:"status"`
}

// applyColorFunc defines a type for applying color to a string.
type applyColorFunc func(a ...interface{}) string

var (
	applyGreen    applyColorFunc = color.New(color.FgGreen).SprintFunc()
	applyRed      applyColorFunc = color.New(color.FgRed).SprintFunc()
	applyYellow	  applyColorFunc = color.New(color.FgYellow).SprintFunc()
	applyBoldRed  applyColorFunc = color.New(color.FgRed, color.Bold).SprintFunc()
	applyMagenta  applyColorFunc = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// logStatus prints a colored log with its HTTP status code
func logStatus(r *http.Request, status int , message string) {
	var applyColor applyColorFunc

	// Set color based on HTTP status code range
	switch {
	case status >= 500 && status < 600:
		// HTTP 5xx: Server errors
		applyColor = applyBoldRed
	case status >= 200 && status < 300:
		// HTTP 2xx: Success
		applyColor = applyGreen
	case status >= 300 && status < 400:
		// HTTP 3xx: Redirection
		applyColor = applyYellow
	case status >= 400 && status < 500:
		// HTTP 4xx: Client errors
		applyColor = applyRed
	default:
		// Unexpected errors
		applyColor = applyMagenta
	}

	// Print log message with appropriate color.
	log.Print(applyColor(fmt.Sprintf("[%d] [%s] [%s] %s", status, r.Method, r.RequestURI, message)))
}

// NotAvailable Notifies the client that the resource does not exist
func NotAvailable(w http.ResponseWriter, r *http.Request) {
	errMsg := "Endpoint not found"
	http.Error(w, errMsg, http.StatusNotFound)
	logStatus(r, http.StatusNotFound, errMsg)
}

// CreateTransaction creates a transaction inside the database and returns the transaction url
func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	// Get the Authorization header value from the request
	if auth := r.Header.Get("Authorization"); auth != fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")) {
		errMsg := "Unauthorized"
		logStatus(r, http.StatusUnauthorized, errMsg)
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	// Parse JSON request body into the TransactionInput struct
	var transactionInput transactions.TransactionInput
	err := json.NewDecoder(r.Body).Decode(&transactionInput)
	if err != nil {
		errMsg := "Invalid JSON input"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Create a new transaction from the TransactionInput
	transaction := transactions.Create(transactionInput)
	id, err := transactions.Insert(r.Context(), &transaction)
	if err != nil {
		errMsg := "Failed to insert transaction"
		logStatus(r, http.StatusInternalServerError, errMsg)
    	http.Error(w, errMsg, http.StatusInternalServerError)
    	return
	}

	// Construct the transaction URL
	transactionURL := fmt.Sprintf("http://%s/transaction/%s", r.Host, id.Hex())

	// Return the transaction URL in the response
	logStatus(r, http.StatusCreated, "Transaction Initialized")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": transactionURL})
	return
}

// redirectUser responds to the request with a redirect url
func redirectUser(w http.ResponseWriter, url string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusSeeOther)
	json.NewEncoder(w).Encode(map[string]interface{}{"url": url})
}

// PostTransaction handles the payment and callback
func PostTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := mux.Vars(r)["transaction_id"]

	// Parse the transaction_id to an objectID
	id, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		errMsg := "Incorrect URI"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Get the transaction
	transaction, err := transactions.GetByID(r.Context(), id)
	if err != nil {
		errMsg := "Failed to get transaction"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Delete the transaction after the function completes
	defer func() {
		err = transactions.Delete(r.Context(), transaction.ID)
		if err != nil {
			log.Printf("[Warning] failed to delete transaction with ID %s from the database: %v", transaction.ID.Hex(), err)
		}
	}()

	// Create a StatusData instance
    statusData := StatusData{
        Status: "Success",
    }

	// Marshal the statusData struct into JSON
	data, err := json.Marshal(statusData)
	if err != nil {
		errMsg := "Failed to create update request body"
		logStatus(r, http.StatusInternalServerError, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Create a new request with the desired method, URL, and body
    req, err := http.NewRequest("POST", transaction.WebhookURL, bytes.NewBuffer(data))
    if err != nil {
		errMsg := "Failed to construct update request"
		logStatus(r, http.StatusInternalServerError, errMsg)
        http.Error(w, errMsg, http.StatusInternalServerError)
        return
    }

    // Set the request headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", transaction.WebhookKey))

    // Create a custom http.Client that skips TLS verification
    httpClient := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }

    // Send the request using the custom httpClient
    response, err := httpClient.Do(req)
    if err != nil {
		errMsg := "Could not reach webhook"
		logStatus(r, http.StatusSeeOther, errMsg)
        redirectUser(w, transaction.RedirectURL)
        return
    }
    defer response.Body.Close()

	// Check if the response was successful
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		// Return the redirect URL in the response
		logStatus(r, http.StatusSeeOther, "Transaction Completed")

		// Redirect the user to the redirection URL
		redirectUser(w, transaction.RedirectURL)
		return
	}

	// Handle non-successful responses
	errMsg := fmt.Sprintf("Source returned error")
	logStatus(r, response.StatusCode, errMsg)

	// Redirect the user to the redirection URL
	redirectUser(w, transaction.RedirectURL)
	return
}

// GetTransactionHTML renders the HTML for the transaction page
func GetTransactionHTML(w http.ResponseWriter, r *http.Request) {
	transactionID := mux.Vars(r)["transaction_id"]

	// Parse the transaction_id to an objectID
	id, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		errMsg := "Incorrect URI"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Get the transaction
	transaction, err := transactions.GetByID(r.Context(), id)
	if err != nil {
		errMsg := "Failed to get transaction"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Setup the transaction page variables
	data := struct {
		Amount float64
		ID     string
	}{
		Amount: transaction.Amount,
		ID:     transaction.ID.Hex(),
	}

	// Set the Content-Type header to specify that the response is HTML
	w.Header().Set("Content-Type", "text/html")

	// Render the transaction page
	err = templates.RenderHTML(w, "transaction.html", data)
	if err != nil {
		errMsg := "Failed to render HTML template"
		logStatus(r, http.StatusInternalServerError, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Log the successful request
	logStatus(r, http.StatusOK, "Served HTML template")
	return
}

// GetTransactionJS renders the JS for the transaction page
func GetTransactionJS(w http.ResponseWriter, r *http.Request) {
	transactionID := mux.Vars(r)["transaction_id"]
	
	// Parse the transaction_id to an objectID
	id, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		errMsg := "Incorrect URI"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Get the transaction
	transaction, err := transactions.GetByID(r.Context(), id)
	if err != nil {
		errMsg := "Failed to get transaction"
		logStatus(r, http.StatusBadRequest, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Setup the transaction javascript variables
	data := struct {
		ID string
	}{
		ID: transaction.ID.Hex(),
	}

	// Set the Content-Type header to specify that the response is JavaScript
	w.Header().Set("Content-Type", "application/javascript")

	// Render the transaction javascript
	err = templates.RenderJS(w, "transaction.js", data)
	if err != nil {
		errMsg := "Failed to render JS template"
		logStatus(r, http.StatusInternalServerError, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Log the successful request
	logStatus(r, http.StatusOK, "Served Javascript template")
	return
}

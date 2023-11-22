package transactions

import (
	"context"
	"errors"
	"dev-payment-gate/utils/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction represents the BSON data stored in the transaction collection
type Transaction struct {
	ID			primitive.ObjectID `bson:"_id,omitempty"`
	Amount		float64			   `bson:"amount"`
	WebhookURL	string			   `bson:"webhook_url"`
	WebhookKey	string			   `bson:"webhook_key"`
	RedirectURL	string			   `bson:"redirect_url"`
	Timestamp	time.Time		   `bson:"timestamp"`
}

// TransactionInput represents the JSON data received to initialize a transaction
type TransactionInput struct {
	Amount		float64	`json:"amount"`
	WebhookURL	string	`json:"webhook_url"`
	WebhookKey	string	`json:"webhook_key"`
	RedirectURL	string	`json:"redirect_url"`
}

// Create initializes a new transaction object
func Create(input TransactionInput) Transaction {
	return Transaction{
		Amount:      input.Amount,
		WebhookURL:  input.WebhookURL,
		WebhookKey:  input.WebhookKey,
		RedirectURL: input.RedirectURL,
		Timestamp:   time.Now(),
	}
}

// Insert stores a transaction into the database and returns its object id
func Insert(ctx context.Context, transaction *Transaction) (*primitive.ObjectID, error) {
	// Setup the database request
	collection := database.GetCollection("transactions")

	// Insert the transaction into the collection "transactions"
	insertOneResult , err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, err
	}

	// Assert the InsertedID as a primitive.ObjectID
	id, ok := insertOneResult.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("Failed to assert InsertedID as primitive.ObjectID")
	}

	// If the assertion succeeds, id contains the ObjectID value
	return &id, nil
}

// GetByID retrieves a transaction from the database using the ID
func GetByID(ctx context.Context, id primitive.ObjectID) (*Transaction, error) {
	// Setup the database request
	collection := database.GetCollection("transactions")
	filter := bson.M{"_id": id}

	// Get the transaction from the collection "transactions"
	var transaction Transaction
	err := collection.FindOne(ctx, filter).Decode(&transaction)
	if err != nil {
		return nil, err
	}
	
	// If no error was received, return the transaction
	return &transaction, nil
}

// Delete removes a transaction from the database
func Delete(ctx context.Context, id primitive.ObjectID) (error) {
	// Setup the database request
	collection := database.GetCollection("transactions")
	filter := bson.M{"_id": id}

	// Delete the transaction from the database
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// Check if a transaction was deleted from the database
	if deleteResult.DeletedCount == 0 {
		return errors.New("There was not a transaction to be deleted")
	}

	// If an entry was deleted, return without an error
	return nil
}

package filter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"filter_unassigned_addresses/util"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FilterUnassignedAddresses() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	postgresURI := os.Getenv("POSTGRES_URI")
	ctx := context.Background()

	generatedAddressesFile := "data/generated_addresses.json"
	assignedAddressesFile := "data/assigned_addresses.json"
	unassignedAddressesFile := "data/unassigned_addresses.json"

	var generatedAddresses []util.MongoAddress
	var assignedAddresses []string
	var unassignedAddresses []util.UnassignedAddress

	// Check if generated addresses file exists
	if _, err := os.Stat(generatedAddressesFile); os.IsNotExist(err) {
		// Connect to MongoDB
		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
		defer mongoClient.Disconnect(ctx)

		// Query generated addresses
		mongoColl := mongoClient.Database(mongoDB).Collection("addresses")
		filter := bson.D{
			{Key: "client", Value: "SENDCASH"},
			{Key: "currencyISO", Value: "BTC"},
		}
		cursor, err := mongoColl.Find(ctx, filter)
		if err != nil {
			log.Fatalf("Failed to query MongoDB: %v", err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var address util.MongoAddress
			if err := cursor.Decode(&address); err != nil {
				log.Fatalf("Failed to decode address: %v", err)
			}
			generatedAddresses = append(generatedAddresses, address)
		}

		if err := cursor.Err(); err != nil {
			log.Fatalf("Cursor error: %v", err)
		}

		// Write generated addresses to file
		data, err := json.Marshal(generatedAddresses)
		if err != nil {
			log.Fatalf("Failed to marshal generated addresses: %v", err)
		}
		if err := ioutil.WriteFile(generatedAddressesFile, data, 0644); err != nil {
			log.Fatalf("Failed to write generated addresses file: %v", err)
		}
	} else {
		// Read generated addresses from file
		data, err := ioutil.ReadFile(generatedAddressesFile)
		if err != nil {
			log.Fatalf("Failed to read generated addresses file: %v", err)
		}
		if err := json.Unmarshal(data, &generatedAddresses); err != nil {
			log.Fatalf("Failed to unmarshal generated addresses: %v", err)
		}
	}

	// Check if assigned addresses file exists
	// It's recommended that you run rails query to fetch all assigned addresses
	if _, err := os.Stat(assignedAddressesFile); os.IsNotExist(err) {
		// Connect to PostgreSQL
		config, err := pgxpool.ParseConfig(postgresURI)
		if err != nil {
			log.Fatalf("Failed to parse PostgreSQL connection string: %v", err)
		}
		config.ConnConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		postgresConn, err := pgxpool.ConnectConfig(ctx, config)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer postgresConn.Close()

		// Query assigned addresses
		rows, err := postgresConn.Query(ctx, "SELECT value FROM addresses WHERE deprecated = false AND cryptocurrency = 'bitcoin'")
		if err != nil {
			log.Fatalf("Failed to query PostgreSQL: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				log.Fatalf("Failed to scan row: %v", err)
			}
			assignedAddresses = append(assignedAddresses, value)
		}

		// Write assigned addresses to file
		data, err := json.Marshal(assignedAddresses)
		if err != nil {
			log.Fatalf("Failed to marshal assigned addresses: %v", err)
		}
		if err := ioutil.WriteFile(assignedAddressesFile, data, 0644); err != nil {
			log.Fatalf("Failed to write assigned addresses file: %v", err)
		}
	} else {
		// Read assigned addresses from file
		data, err := ioutil.ReadFile(assignedAddressesFile)
		if err != nil {
			log.Fatalf("Failed to read assigned addresses file: %v", err)
		}
		if err := json.Unmarshal(data, &assignedAddresses); err != nil {
			log.Fatalf("Failed to unmarshal assigned addresses: %v", err)
		}
	}

	// Convert assignedAddresses slice to a map for easy lookup
	assignedAddressesMap := make(map[string]struct{})
	for _, address := range assignedAddresses {
		assignedAddressesMap[address] = struct{}{}
	}

	// Determine unassigned addresses
	currentTime := time.Now()
	for _, address := range generatedAddresses {
		if _, assigned := assignedAddressesMap[address.Value]; !assigned {
			id, err := primitive.ObjectIDFromHex(address.ID)

			if err != nil {
				log.Fatalf("Failed to convert string to ObjectID: %v", err)
			}

			unassignedAddresses = append(unassignedAddresses, util.UnassignedAddress{
				ID:          id,
				Value:       address.Value,
				Type:        address.Type,
				CurrencyISO: address.CurrencyISO,
				Client:      address.Client,
				GeneratedAt: primitive.NewDateTimeFromTime(address.CreatedAt),
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			})
		}
	}

	// Write unassigned addresses to file
	data, err := json.Marshal(unassignedAddresses)
	if err != nil {
		log.Fatalf("Failed to marshal unassigned addresses: %v", err)
	}
	if err := ioutil.WriteFile(unassignedAddressesFile, data, 0644); err != nil {
		log.Fatalf("Failed to write unassigned addresses file: %v", err)
	}

	fmt.Println("Successfully filtered and saved unassigned addresses.")
}

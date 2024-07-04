package persist

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"filter_unassigned_addresses/util"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PersistUnassigned() {
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	unassignedAddressesFile := "data/unassigned_addresses.json"

	ctx := context.Background()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	fmt.Println("Mongo connected")
	fmt.Println("Mongo connected")
	defer mongoClient.Disconnect(ctx)

	// Read unassigned addresses from file
	data, err := ioutil.ReadFile(unassignedAddressesFile)
	if err != nil {
		log.Fatalf("Failed to read unassigned addresses file: %v", err)
	}
	fmt.Println("unassigned addresses file loaded")

	var unassignedAddresses []util.UnassignedAddress
	if err := json.Unmarshal(data, &unassignedAddresses); err != nil {
		log.Fatalf("Failed to unmarshal unassigned addresses: %v", err)
	}
	for i := range unassignedAddresses {
		unassignedAddresses[i].CreatedAt = time.Now()
		unassignedAddresses[i].UpdatedAt = time.Now()
	}

	// Insert the records into the MongoDB collection
	mongoColl := mongoClient.Database(mongoDB).Collection("unassigned_addresses")
	var docs []interface{}
	for _, address := range unassignedAddresses {
		docs = append(docs, address)
	}

	_, err = mongoColl.InsertMany(ctx, docs)
	if err != nil {
		log.Fatalf("Failed to insert unassigned addresses into MongoDB: %v", err)
	}

	fmt.Println("Successfully persisted unassigned addresses to MongoDB")
}

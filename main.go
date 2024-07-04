package main

import (
	"filter_unassigned_addresses/filter"
	"filter_unassigned_addresses/persist"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Filter unassigned addresses and write to file
	filter.FilterUnassignedAddresses()

	// Persist unassigned addresses to MongoDB
	persist.PersistUnassigned()
}

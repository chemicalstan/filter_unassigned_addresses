package util

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoAddress struct {
	ID          string    `bson:"_id"`
	Value       string    `bson:"value"`
	Type        string    `bson:"type"`
	CurrencyISO string    `bson:"currencyISO"`
	Client      string    `bson:"client"`
	CreatedAt   time.Time `bson:"createdAt"`
}

type UnassignedAddress struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Value       string             `json:"value" bson:"value"`
	Type        string             `json:"type" bson:"type"`
	CurrencyISO string             `json:"currencyISO" bson:"currencyISO"`
	Client      string             `json:"client" bson:"client"`
	GeneratedAt primitive.DateTime `json:"generatedAt" bson:"generatedAt"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

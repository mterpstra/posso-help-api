package chat

import (
	"context"
	"log"
	"posso-help/internal/db"
	"go.mongodb.org/mongo-driver/bson"
)

const MessagesCollection = "messages"

type ParsedMessage struct {
	Account     string `bson:"account" json:"account"`
	PhoneNumber string `bson:"phone" json:"phone"`
	Name        string `bson:"name" json:"name"`
	Date        string `bson:"date" json:"date"`
	RawMessage  string `bson:"raw_message" json:"raw_message"`
	MessageType string `bson:"message_type" json:"message_type"`
}

func SaveParsedMessage(bmv *BaseMessageValues, rawMessage string, messageType string) error {
	collection := db.GetCollection(MessagesCollection)

	doc := bson.M{
		"account":      bmv.Account,
		"phone":        bmv.PhoneNumber,
		"name":         bmv.Name,
		"date":         bmv.Date,
		"raw_message":  rawMessage,
		"message_type": messageType,
	}

	_, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Printf("Error saving parsed message: %v\n", err)
		return err
	}

	log.Printf("Saved parsed message to messages collection: type=%s\n", messageType)
	return nil
}

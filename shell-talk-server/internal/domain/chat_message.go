package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessage represents a single message in a conversation, stored in MongoDB.
type ChatMessage struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	ConversationID string             `bson:"conversation_id"`
	SenderID       string             `bson:"sender_id"`
	SenderNickname string             `bson:"sender_nickname"`
	Content        string             `bson:"content"`
	Timestamp      time.Time          `bson:"timestamp"`
}

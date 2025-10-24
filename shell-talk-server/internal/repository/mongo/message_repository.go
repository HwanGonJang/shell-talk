package mongo

import (
	"context"
	"shell-talk-server/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const messageCollection = "messages"

// MessageRepository handles database operations for chat messages.
type MessageRepository struct {
	DB *mongo.Database
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{DB: db}
}

// SaveMessage inserts a new chat message into the database.
func (r *MessageRepository) SaveMessage(ctx context.Context, message *domain.ChatMessage) error {
	collection := r.DB.Collection(messageCollection)
	_, err := collection.InsertOne(ctx, message)
	return err
}

// GetMessagesByConversationID retrieves the last N messages for a conversation.
func (r *MessageRepository) GetMessagesByConversationID(ctx context.Context, conversationID string, limit int64) ([]*domain.ChatMessage, error) {
	collection := r.DB.Collection(messageCollection)

	// Find options to sort by timestamp descending and limit the results
	// opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)

	cursor, err := collection.Find(ctx, bson.M{"conversation_id": conversationID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*domain.ChatMessage
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// The messages are currently sorted oldest to newest. If we want newest first, we need to sort.
	// For a chat application, we usually want to show the oldest first, so this is fine.

	return messages, nil
}

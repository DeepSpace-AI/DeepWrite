package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/deepwrite/writing-service/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ContentRepository interface {
	Create(ctx context.Context, content *models.DocumentContent) error
	GetByDocumentID(ctx context.Context, documentID string) (*models.DocumentContent, error)
	Update(ctx context.Context, documentID string, content *models.DocumentContent) error
	Delete(ctx context.Context, documentID string) error
}

type MongoContentRepository struct {
	collection *mongo.Collection
}

func NewMongoContentRepository(db *mongo.Database) *MongoContentRepository {
	return &MongoContentRepository{
		collection: db.Collection("document_contents"),
	}
}

func (r *MongoContentRepository) Create(ctx context.Context, content *models.DocumentContent) error {
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, content)
	return err
}

func (r *MongoContentRepository) GetByDocumentID(ctx context.Context, documentID string) (*models.DocumentContent, error) {
	var content models.DocumentContent
	err := r.collection.FindOne(ctx, bson.M{"document_id": documentID}).Decode(&content)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *MongoContentRepository) Update(ctx context.Context, documentID string, content *models.DocumentContent) error {
	content.UpdatedAt = time.Now()
	filter := bson.M{"document_id": documentID}
	update := bson.M{
		"$set": bson.M{
			"title":     content.Title,
			"content":   content.Content,
			"metadata":  content.Metadata,
			"citations": content.Citations,
			"updated_at": content.UpdatedAt,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoContentRepository) Delete(ctx context.Context, documentID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"document_id": documentID})
	return err
}

type SnapshotRepository interface {
	Create(ctx context.Context, documentID string, content *models.DocumentContent) (string, error)
	GetByDocumentID(ctx context.Context, documentID string, page, limit int) ([]*models.DocumentContent, int, error)
	GetLatest(ctx context.Context, documentID string) (*models.DocumentContent, error)
}

type MongoSnapshotRepository struct {
	collection *mongo.Collection
}

func NewMongoSnapshotRepository(db *mongo.Database) *MongoSnapshotRepository {
	return &MongoSnapshotRepository{
		collection: db.Collection("document_snapshots"),
	}
}

func (r *MongoSnapshotRepository) Create(ctx context.Context, documentID string, content *models.DocumentContent) (string, error) {
	snapshot := bson.M{
		"document_id": documentID,
		"title":       content.Title,
		"content":     content.Content,
		"metadata":    content.Metadata,
		"citations":   content.Citations,
		"created_at":  time.Now(),
	}
	result, err := r.collection.InsertOne(ctx, snapshot)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result.InsertedID), nil
}

func (r *MongoSnapshotRepository) GetByDocumentID(ctx context.Context, documentID string, page, limit int) ([]*models.DocumentContent, int, error) {
	filter := bson.M{"document_id": documentID}
	
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var snapshots []*models.DocumentContent
	if err := cursor.All(ctx, &snapshots); err != nil {
		return nil, 0, err
	}

	return snapshots, int(count), nil
}

func (r *MongoSnapshotRepository) GetLatest(ctx context.Context, documentID string) (*models.DocumentContent, error) {
	var content models.DocumentContent
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	err := r.collection.FindOne(ctx, bson.M{"document_id": documentID}, opts).Decode(&content)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &content, nil
}

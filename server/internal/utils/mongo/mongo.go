package mongo

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func StringToObjectID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, fmt.Errorf("empty id")
	}
	return primitive.ObjectIDFromHex(id)
}

func IsValidObjectID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}

func GenerateTestID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func FixedTestID() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	return id
}

func MustObjectID(t *testing.T, hex string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		t.Fatalf("Failed to create ObjectID from %s: %v", hex, err)
	}
	return id
}

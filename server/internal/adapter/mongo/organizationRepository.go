package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"semki/internal/model"
	"semki/pkg/clients"
)

type IOrganizationRepository interface {
	CreateOrganization(ctx context.Context, organization model.Organization) error
	GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error)
	GetOrganizationByTitle(ctx context.Context, email string) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error
	DeleteOrganization(ctx context.Context, id primitive.ObjectID) error
	PatchOrganization(ctx context.Context, orgID primitive.ObjectID, updates bson.M) error
	AddLevel(ctx context.Context, orgID primitive.ObjectID, level model.Level) error
	UpdateLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID, updates bson.M) error
	DeleteLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID) error
	AddTeam(ctx context.Context, orgID primitive.ObjectID, team model.Team) error
	UpdateTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID, updates bson.M) error
	DeleteTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID) error
	AddLocation(ctx context.Context, orgID primitive.ObjectID, location model.Location) error
	DeleteLocation(ctx context.Context, orgID primitive.ObjectID, locationID primitive.ObjectID) error
}

type organizationRepository struct {
	client *clients.MongoDb
}

func NewOrganizationRepository(client *clients.MongoDb) IOrganizationRepository {
	return &organizationRepository{client}
}

//region Organizations

func (r *organizationRepository) CreateOrganization(ctx context.Context, organization model.Organization) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	_, err := coll.InsertOne(ctx, organization)
	return err
}

func (r *organizationRepository) GetOrganizationByID(ctx context.Context, id primitive.ObjectID) (*model.Organization, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	var organization model.Organization
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &organization, nil
}

func (r *organizationRepository) GetOrganizationByTitle(ctx context.Context, title string) (*model.Organization, error) {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	var organization model.Organization
	err := coll.FindOne(ctx, bson.M{"title": title}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &organization, nil
}

func (r *organizationRepository) UpdateOrganization(ctx context.Context, id primitive.ObjectID, organization model.Organization) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	_, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, organization)
	return err
}

func (r *organizationRepository) DeleteOrganization(ctx context.Context, id primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)
	usersColl := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Users)

	_, err := usersColl.DeleteMany(ctx, bson.M{"organization_id": id})
	if err != nil {
		return err
	}

	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// PatchOrganization updates only provided fields
func (r *organizationRepository) PatchOrganization(ctx context.Context, orgID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{"$set": updates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Team methods

func (r *organizationRepository) AddTeam(ctx context.Context, orgID primitive.ObjectID, team model.Team) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.teams": team},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *organizationRepository) UpdateTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	// Build update document for positional operator
	setUpdates := bson.M{}
	for key, value := range updates {
		setUpdates["semantic.teams.$."+key] = value
	}

	filter := bson.M{
		"_id":                orgID,
		"semantic.teams._id": teamID,
	}
	update := bson.M{"$set": setUpdates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *organizationRepository) DeleteTeam(ctx context.Context, orgID primitive.ObjectID, teamID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.teams": bson.M{"_id": teamID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Level methods

func (r *organizationRepository) AddLevel(ctx context.Context, orgID primitive.ObjectID, level model.Level) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.levels": level},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *organizationRepository) UpdateLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID, updates bson.M) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	setUpdates := bson.M{}
	for key, value := range updates {
		setUpdates["semantic.levels.$."+key] = value
	}

	filter := bson.M{
		"_id":                 orgID,
		"semantic.levels._id": levelID,
	}
	update := bson.M{"$set": setUpdates}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *organizationRepository) DeleteLevel(ctx context.Context, orgID primitive.ObjectID, levelID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.levels": bson.M{"_id": levelID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Location methods

func (r *organizationRepository) AddLocation(ctx context.Context, orgID primitive.ObjectID, location model.Location) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$push": bson.M{"semantic.locations": location},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *organizationRepository) DeleteLocation(ctx context.Context, orgID primitive.ObjectID, locationID primitive.ObjectID) error {
	coll := r.client.Client.Database(r.client.Database).Collection(r.client.Collections.Organizations)

	filter := bson.M{"_id": orgID}
	update := bson.M{
		"$pull": bson.M{
			"semantic.locations": bson.M{"_id": locationID},
		},
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

//endregion

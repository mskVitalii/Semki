package mongo_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"semki/internal/adapter/mongo"
	"semki/internal/model"
	"semki/internal/utils/config"
	"semki/pkg/clients"
	"semki/pkg/lib"
	"semki/pkg/telemetry"
	"testing"
)

func TestSoftDeleteUser(t *testing.T) {
	// Arrange
	db, repo, user, ctx := arrangeMongo(t)

	// Act: Delete user
	err := repo.DeleteUser(ctx, user.ID)
	assert.NoError(t, err)

	// Assert: Check status "deleted"
	retrievedUser, err := repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.UserStatuses.DELETED, retrievedUser.Status)

	cleanup(t, db, ctx)
}

func TestUserTypes(t *testing.T) {
	// Arrange
	db, repo, user, ctx := arrangeMongo(t)

	// Act =================== regular
	// regular user can add only 1 home & 1 favourite
	err := repo.UpdateUser(ctx, user.ID, model.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		Status:   user.Status,
	})
	assert.NoError(t, err)
	_, err = repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	//assert.Equal(t, len(userAfterAddingHome.Homes), 1)
	//assert.Equal(t, len(userAfterAddingHome.Favourites), 1)

	err = repo.UpdateUser(ctx, user.ID, model.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		Status:   user.Status,
	})
	assert.NoError(t, err)
	_, err = repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	//assert.Equal(t, len(userAfterAddingHome.Homes), 1)
	//assert.Equal(t, len(userAfterAddingHome.Favourites), 1)

	// Act =================== BUSINESS
	err = repo.UpdateUser(ctx, user.ID, model.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		Status:   user.Status,
	})
	assert.NoError(t, err)

	// BUSINESS user can add more than 1 home & more than 1 favourite place
	err = repo.UpdateUser(ctx, user.ID, model.User{
		ID:        user.ID,
		Email:     user.Email,
		Password:  user.Password,
		Providers: []model.UserProvider{model.UserProviders.Google},
		Status:    user.Status,
	})
	assert.NoError(t, err)
	_, err = repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	//assert.Equal(t, len(userAfterAddingHome.Homes), 2)
	//assert.Equal(t, len(userAfterAddingHome.Favourites), 2)

	cleanup(t, db, ctx)
}

func arrangeMongo(t *testing.T) (*clients.MongoDb, mongo.IUserRepository, model.User, context.Context) {
	cfg := config.GetConfig("../../../")
	telemetry.SetupLogger(cfg)

	// Arrange
	db, err := mongo.SetupMongo(&cfg.Mongo)
	if err != nil {
		t.Fatal(err)
	}

	repo := mongo.NewUserRepository(cfg, &clients.MongoDb{
		Database:    "test_db",
		Client:      db.Client,
		Collections: clients.MongoCollectionsNames})

	user := model.User{
		ID:       primitive.NewObjectID(),
		Email:    "test@example.com",
		Password: lib.HashPassword("password"),
		Status:   model.UserStatuses.ACTIVE,
	}

	ctx := context.TODO()
	err = repo.CreateUser(ctx, &user)
	assert.NoError(t, err)

	_, err = repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)

	return db, repo, user, ctx
}

func cleanup(t *testing.T, db *clients.MongoDb, ctx context.Context) {
	err := db.Client.Database("test_db").Drop(ctx)
	assert.NoError(t, err)
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		assert.NoError(t, err)
	}(telemetry.Log)
}

package sessions

import (
	"context"
	"os"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v8"
	"github.com/tsaron/anansi"
	"github.com/tsaron/anansi/postgres"
	"github.com/tsaron/anansi/tokens"
	"syreclabs.com/go/faker"
	"tsaron.com/godview-starter/pkg/config"
	"tsaron.com/godview-starter/pkg/users"
	"tsaron.com/godview-starter/pkg/workspaces"
)

var testDB *pg.DB
var store *tokens.Store
var mem *redis.Client

func afterEach(t *testing.T) {
	if err := postgres.CleanUpTables(testDB, "users", "workspaces"); err != nil {
		t.Fatal(err)
	}

	if _, err := mem.FlushDB(context.TODO()).Result(); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	var env config.Env
	if err = anansi.LoadEnv(&env); err != nil {
		panic(err)
	}

	log := anansi.NewLogger(env.Name)

	if testDB, err = config.SetupDB(env); err != nil {
		panic(err)
	}
	log.Info().Msg("Successfully connected to postgres")

	if mem, err = config.SetupRedis(context.TODO(), env); err != nil {
		panic(err)
	}
	store = tokens.NewStore(mem, env.Secret)

	defer os.Exit(m.Run())

	if err := testDB.Close(); err != nil {
		log.Err(err).Msg("Failed to disconnect from postgres cleanly")
	}

	if err := mem.Close(); err != nil {
		panic(err)
	}
}

func TestStoreCreate(t *testing.T) {
	defer afterEach(t)

	ctx := context.TODO()
	wkRepo := workspaces.NewRepo(testDB)
	sessions := NewStore(store, wkRepo)

	wk, err := wkRepo.Create(ctx, faker.Company().Name(), faker.Internet().Email())
	if err != nil {
		t.Fatal(err)
	}

	user := &users.User{
		ID:           1,
		FirstName:    faker.Name().FirstName(),
		LastName:     faker.Name().LastName(),
		Role:         faker.RandomChoice([]string{"admin", "member"}),
		EmailAddress: faker.Internet().Email(),
		Workspace:    wk.ID,
	}

	session, err := sessions.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}

	loaded := new(Session)
	if err := store.Peek(ctx, session.SessionKey, loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.User != session.User {
		t.Errorf("Expected loaded session to be the user %d, got user %d", session.User, loaded.User)
	}
}

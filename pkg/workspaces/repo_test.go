package workspaces

import (
	"context"
	"os"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/tsaron/anansi"
	"github.com/tsaron/anansi/postgres"
	"syreclabs.com/go/faker"
	"tsaron.com/godview-starter/pkg/config"
)

var testDB *pg.DB

func afterEach(t *testing.T) {
	if err := postgres.CleanUpTables(testDB, "workspaces"); err != nil {
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

	code := m.Run()

	if err := testDB.Close(); err != nil {
		log.Err(err).Msg("Failed to disconnect from postgres cleanly")
	}

	os.Exit(code)
}

func TestRepoGetByID(t *testing.T) {
	repo := NewRepo(testDB)
	ctx := context.TODO()

	t.Run("returns a workspace based on its ID", func(t *testing.T) {
		defer afterEach(t)

		wk, err := repo.Create(ctx, faker.Company().Name(), faker.Internet().Email())
		if err != nil {
			t.Fatal(err)
		}

		wk2, err := repo.Get(ctx, wk.ID)
		if err != nil {
			t.Fatal(err)
		}

		if wk2.ID != wk.ID {
			t.Errorf("Expected loaded workspace(%d) to be the same as created workspace(%d)", wk2.ID, wk.ID)
		}
	})

	t.Run("return nil if the workspace doesn't exist", func(t *testing.T) {
		defer afterEach(t)

		wk, err := repo.Get(ctx, uint(faker.RandomInt(1, 20)))
		if err != nil {
			t.Fatal(err)
		}

		if wk != nil {
			t.Errorf("Expected loaded workspace to be the nil found %v", *wk)
		}
	})
}

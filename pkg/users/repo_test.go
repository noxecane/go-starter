package users

import (
	"context"
	"os"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/tsaron/anansi"
	"github.com/tsaron/anansi/postgres"
	"syreclabs.com/go/faker"
	"tsaron.com/godview-starter/pkg/config"
	"tsaron.com/godview-starter/pkg/workspaces"
)

var testDB *pg.DB

func afterEach(t *testing.T) {
	if err := postgres.CleanUpTables(testDB, "users", "workspaces"); err != nil {
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

func TestRepoCreate(t *testing.T) {
	defer afterEach(t)

	repo := NewRepo(testDB)
	ctx := context.TODO()

	wkRepo := workspaces.NewRepo(testDB)
	wk, err := wkRepo.Create(ctx, faker.Company().Name(), faker.Internet().Email())
	if err != nil {
		t.Fatal(err)
	}

	req := UserRequest{faker.Internet().Email(), faker.Fetch("name.title.job")}
	_, err = repo.Create(ctx, wk.ID, req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Create(ctx, wk.ID, req)
	if err == nil {
		t.Fatalf("Expected duplicate create to fail")
	}

	switch err.(type) {
	case ErrEmail:
		// no op
	default:
		t.Errorf("Expected error to be of type errEmail, got %T", err)
	}
}

func TestRepoCreateMany(t *testing.T) {
	defer afterEach(t)

	repo := NewRepo(testDB)
	ctx := context.TODO()

	wkRepo := workspaces.NewRepo(testDB)
	wk, err := wkRepo.Create(ctx, faker.Company().Name(), faker.Internet().Email())
	if err != nil {
		t.Fatal(err)
	}

	req := UserRequest{faker.Internet().Email(), faker.Fetch("name.title.job")}
	_, err = repo.Create(ctx, wk.ID, req)
	if err != nil {
		t.Fatal(err)
	}

	reqs := []UserRequest{
		{faker.Internet().Email(), faker.Fetch("name.title.job")},
		req,
	}
	_, err = repo.CreateMany(ctx, wk.ID, reqs)
	if err == nil {
		t.Fatalf("Expected duplicate create to fail")
	}

	switch err.(type) {
	case ErrEmail:
		// no op
	default:
		t.Errorf("Expected error to be of type ErrEmail, got %T: %v", err, err)
	}
}

func TestRepoRegister(t *testing.T) {
	defer afterEach(t)

	repo := NewRepo(testDB)
	ctx := context.TODO()

	wkRepo := workspaces.NewRepo(testDB)
	wk, err := wkRepo.Create(ctx, faker.Company().Name(), faker.Internet().Email())
	if err != nil {
		t.Fatal(err)
	}

	reqs := []UserRequest{
		{faker.Internet().Email(), faker.Fetch("name.title.job")},
		{faker.Internet().Email(), faker.Fetch("name.title.job")},
	}
	_, err = repo.CreateMany(ctx, wk.ID, reqs)
	if err != nil {
		t.Fatal(err)
	}

	reg := Registration{
		faker.Name().FirstName(),
		faker.Name().LastName(),
		faker.Lorem().Word(),
		faker.PhoneNumber().PhoneNumber(),
	}

	_, err = repo.Register(ctx, reqs[0].EmailAddress, reg)
	if err != nil {
		t.Fatal(err)
	}

	reg2 := Registration{
		faker.Name().FirstName(),
		faker.Name().LastName(),
		faker.Lorem().Word(),
		reg.PhoneNumber,
	}
	_, err = repo.Register(ctx, reqs[1].EmailAddress, reg2)
	if err != ErrExistingPhoneNumber {
		t.Errorf("Expected registeration with \"%v\", got %v", ErrExistingPhoneNumber, err)
	}
}

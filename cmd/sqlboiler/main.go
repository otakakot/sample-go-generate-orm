package main

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	// postgres driver.
	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/otakakot/sample-go-generate-orm/pkg/sqlboilergen"
)

func main() {
	dsn := cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	// create
	user1 := sqlboilergen.User{
		Name: uuid.NewString(),
	}

	if err := user1.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	// transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	user2 := sqlboilergen.User{
		Name: uuid.NewString(),
	}

	if err := user2.Insert(ctx, tx, boil.Infer()); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	user3 := sqlboilergen.User{
		Name: uuid.NewString(),
	}

	if err := user3.Insert(ctx, tx, boil.Infer()); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	// read
	got, err := sqlboilergen.FindUser(ctx, db, user1.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user: %+v", got))

	gots, err := sqlboilergen.Users(
		sqlboilergen.UserWhere.ID.IN([]string{user1.ID}),
		sqlboilergen.UserWhere.Name.EQ(user1.Name),
	).All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, got := range gots {
		slog.Info(fmt.Sprintf("user: %+v", got))
	}

	// update
	user1.Name = uuid.NewString()

	if _, err := user1.Update(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	if user, err := sqlboilergen.FindUser(ctx, db, user1.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("updated user: %+v", user))
	}

	// delete
	if _, err := user1.Delete(ctx, db); err != nil {
		panic(err)
	}

	if _, err := sqlboilergen.FindUser(ctx, db, user1.ID); err != sql.ErrNoRows {
		panic(err)
	}
}

package main

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"

	// postgres driver.
	_ "github.com/lib/pq"

	"github.com/otakakot/sample-go-generate-orm/pkg/xogen"
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
	user1 := &xogen.User{
		ID:   uuid.New(),
		Name: uuid.NewString(),
	}

	if err := user1.Insert(ctx, db); err != nil {
		panic(err)
	}

	// transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	user2 := &xogen.User{
		ID:   uuid.New(),
		Name: uuid.NewString(),
	}

	if err := user2.Insert(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	user3 := &xogen.User{
		ID:   uuid.New(),
		Name: uuid.NewString(),
	}

	if err := user3.Insert(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	// read
	got, err := xogen.UserByID(ctx, db, user1.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user: %+v", got))

	// 複数取得はコードが自動生成されない ...

	// update
	user1.Name = uuid.NewString()

	if err := user1.Update(ctx, db); err != nil {
		panic(err)
	}

	if user, err := xogen.UserByID(ctx, db, user1.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("updated user: %+v", user))
	}

	// delete
	if err := user1.Delete(ctx, db); err != nil {
		panic(err)
	}

	if _, err := xogen.UserByID(ctx, db, user1.ID); err != sql.ErrNoRows {
		panic(err)
	}
}

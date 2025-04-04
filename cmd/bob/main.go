package main

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/aarondl/opt/omit"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"

	"github.com/otakakot/sample-go-generate-orm/pkg/bobgen"
)

func main() {
	dsn := cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	db, err := bob.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	ctx := context.Background()

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	// create
	user1, err := bobgen.Users.Insert(&bobgen.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, db)
	if err != nil {
		panic(err)
	}

	// transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	if _, err := bobgen.Users.Insert(&bobgen.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if _, err := bobgen.Users.Insert(&bobgen.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	got, err := bobgen.FindUser(ctx, db, user1.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("got: %+v", got))

	gots, err := bobgen.Users.View.Query(
		bobgen.SelectWhere.Users.ID.EQ(user1.ID),
		bobgen.SelectWhere.Users.Name.EQ(user1.Name),
	).All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, got := range gots {
		slog.Info(fmt.Sprintf("user: %+v", got))
	}

	// update
	if err := user1.Update(ctx, db, &bobgen.UserSetter{
		Name: omit.From(uuid.NewString()),
	}); err != nil {
		panic(err)
	}

	if user, err := bobgen.FindUser(ctx, db, user1.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("user: %+v", user))
	}

	// delete
	if err := user1.Delete(ctx, db); err != nil {
		panic(err)
	}

	if _, err := bobgen.FindUser(ctx, db, user1.ID); err != sql.ErrNoRows {
		panic(err)
	}
}

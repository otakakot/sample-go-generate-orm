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
	"github.com/stephenafamo/bob/dialect/psql/im"

	"github.com/otakakot/sample-go-generate-orm/pkg/bob/models"
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

	// 事前データ準備
	if _, err := db.ExecContext(ctx, "TRUNCATE TABLE users CASCADE"); err != nil {
		panic(err)
	}

	tmp, err := models.Users.Insert(&models.UserSetter{
		Name: omit.From("name"),
	}).One(ctx, db)
	if err != nil {
		panic(err)
	}

	// SELECT
	gots, err := models.Users.View.Query().All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, got := range gots {
		slog.Info(fmt.Sprintf("user: %+v", got))
	}

	// SELECT WHERE
	got, err := models.FindUser(ctx, db, tmp.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user: %+v", got))

	users, err := models.Users.View.Query(
		models.SelectWhere.Users.Name.EQ("name"),
	).All(ctx, db)
	if err != nil {
		panic(err)
	}
	for _, user := range users {
		slog.Info(fmt.Sprintf("user: %+v", user))
	}

	// INSERT
	user, err := models.Users.Insert(&models.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, db)
	if err != nil {
		panic(err)
	}

	// UPDATE
	if err := user.Update(ctx, db, &models.UserSetter{
		Name: omit.From("Updated " + user.Name),
	}); err != nil {
		panic(err)
	}

	updated, err := models.FindUser(ctx, db, tmp.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("updated user: %+v", updated))

	// INSERT ON CONFLICT DO NOTHING
	if _, err := models.Users.Insert(&models.UserSetter{
		ID:   omit.From(updated.ID),
		Name: omit.From("Upserted " + updated.Name),
	}, im.OnConflict("id").DoNothing()).One(ctx, db); err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if user, err := models.FindUser(ctx, db, user.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("upserted user: %+v", user))
	}

	// INSERT ON CONFLICT DO UPDATE
	if _, err := models.Users.Insert(&models.UserSetter{
		ID:   omit.From(user.ID),
		Name: omit.From("Upserted " + updated.Name),
	}, im.OnConflict("id").DoUpdate(
		im.SetExcluded("name"),
	)).One(ctx, db); err != nil {
		panic(err)
	}

	if user, err := models.FindUser(ctx, db, user.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("upserted user: %+v", user))
	}

	// DELETE
	if err := user.Delete(ctx, db); err != nil {
		panic(err)
	}
	if _, err := models.FindUser(ctx, db, user.ID); err != sql.ErrNoRows {
		panic(err)
	}

	exists, err := models.UserExists(ctx, db, user.ID)
	if err != nil {
		panic(err)
	}
	if exists {
		slog.Error("user exists")
	} else {
		slog.Info("user deleted")
	}

	count, err := models.Users.Query().Count(ctx, db)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user count: %d", count))

	// TRANSACTION
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	if _, err := models.Users.Insert(&models.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if _, err := models.Users.Insert(&models.UserSetter{
		Name: omit.From(uuid.NewString()),
	}).One(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	all, err := models.Users.Query().All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, user := range all {
		slog.Info(fmt.Sprintf("user: %+v", user))
	}
}

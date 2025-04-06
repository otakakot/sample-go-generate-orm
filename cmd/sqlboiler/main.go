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

	"github.com/otakakot/sample-go-generate-orm/pkg/sqlboiler/models"
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

	// 事前データ準備
	if _, err := db.QueryContext(ctx, "TRUNCATE TABLE users, posts CASCADE"); err != nil {
		panic(err)
	}

	tmpUser1 := models.User{
		Name: uuid.NewString(),
	}

	if err := tmpUser1.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	tmpPost1 := models.Post{
		UserID:  tmpUser1.ID,
		Title:   "1",
		Content: "1",
	}

	if err := tmpPost1.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	tmpPost2 := models.Post{
		UserID:  tmpUser1.ID,
		Title:   "2",
		Content: "2",
	}

	if err := tmpPost2.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	tmpPost3 := models.Post{
		UserID:  tmpUser1.ID,
		Title:   "3",
		Content: "3",
	}

	if err := tmpPost3.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	// SELECT
	gots, err := models.Users().All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, got := range gots {
		slog.Info(fmt.Sprintf("user: %+v", got))
	}

	// SELECT WHERE
	got, err := models.FindUser(ctx, db, tmpUser1.ID)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user: %+v", got))

	posts, err := models.Posts(
		models.PostWhere.UserID.EQ(got.ID),
	).All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, post := range posts {
		slog.Info(fmt.Sprintf("post: %+v", post))
	}

	// INSERT
	user := models.User{
		Name: uuid.NewString(),
	}

	if err := user.Insert(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	// UPDATE
	user.Name = "Updated " + user.Name
	if _, err := user.Update(ctx, db, boil.Infer()); err != nil {
		panic(err)
	}

	if user, err := models.FindUser(ctx, db, user.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("updated user: %+v", user))
	}

	// INSERT ON CONFLICT DO NOTHING
	user.Name = "Upserted " + user.Name

	if err := user.Upsert(ctx, db, false, nil, boil.Columns{}, boil.Infer()); err != nil {
		panic(err)
	}

	if user, err := models.FindUser(ctx, db, user.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("upserted user: %+v", user))
	}

	// INSERT ON CONFLICT DO UPDATE
	if err := user.Upsert(ctx, db, true, []string{"id"}, boil.Infer(), boil.Infer()); err != nil {
		panic(err)
	}

	if user, err := models.FindUser(ctx, db, user.ID); err != nil {
		panic(err)
	} else {
		slog.Info(fmt.Sprintf("upserted user: %+v", user))
	}

	// DELETE
	if _, err := user.Delete(ctx, db); err != nil {
		panic(err)
	}
	if _, err := models.FindUser(ctx, db, user.ID); err != sql.ErrNoRows {
		panic(err)
	}

	exists, err := user.Exists(ctx, db)
	if err != nil {
		panic(err)
	}
	if exists {
		slog.Error("user exists")
	} else {
		slog.Info("user deleted")
	}

	count, err := models.Users().Count(ctx, db)
	if err != nil {
		panic(err)
	}

	slog.Info(fmt.Sprintf("user count: %d", count))

	// TRANSACTION
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	user1 := models.User{
		Name: uuid.NewString(),
	}
	if err := user1.Insert(ctx, tx, boil.Infer()); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	user2 := models.User{
		Name: uuid.NewString(),
	}
	if err := user2.Insert(ctx, tx, boil.Infer()); err != nil {
		if err := tx.Rollback(); err != nil {
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	users, err := models.Users().All(ctx, db)
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		slog.Info(fmt.Sprintf("user: %+v", user))
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	// postgres driver.
	_ "github.com/lib/pq"

	"github.com/otakakot/sample-go-generate-orm/pkg/sqlboiler"
	"github.com/otakakot/sample-go-generate-orm/pkg/xo"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	now := time.Now()

	// xo
	{
		// create
		user1 := &xo.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := user1.Insert(ctx, db); err != nil {
			panic(err)
		}

		// transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			panic(err)
		}

		user2 := &xo.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := user2.Insert(ctx, tx); err != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}

		user3 := &xo.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
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
		got, err := xo.UserByID(ctx, db, user1.ID)
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

		if user, err := xo.UserByID(ctx, db, user1.ID); err != nil {
			panic(err)
		} else {
			slog.Info(fmt.Sprintf("updated user: %+v", user))
		}

		// delete
		if err := user1.Delete(ctx, db); err != nil {
			panic(err)
		}

		if _, err := xo.UserByID(ctx, db, user1.ID); err != sql.ErrNoRows {
			panic(err)
		}
	}

	// sqlboiler
	{
		// create
		user1 := sqlboiler.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := user1.Insert(ctx, db, boil.Infer()); err != nil {
			panic(err)
		}

		// transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			panic(err)
		}

		user2 := sqlboiler.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := user2.Insert(ctx, tx, boil.Infer()); err != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}

		user3 := sqlboiler.User{
			Name:      uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
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
		got, err := sqlboiler.FindUser(ctx, db, user1.ID)
		if err != nil {
			panic(err)
		}

		slog.Info(fmt.Sprintf("user: %+v", got))

		// select
		// users, err := sqlboiler.Users().All(ctx, db)
		// if err != nil {
		// 	panic(err)
		// }

		// for _, user := range users {
		// 	slog.Info(fmt.Sprintf("user: %+v", user))
		// }

		gots, err := sqlboiler.Users(sqlboiler.UserWhere.ID.IN([]int{user1.ID}), sqlboiler.UserWhere.Name.EQ(user1.Name)).All(ctx, db)
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

		if user, err := sqlboiler.FindUser(ctx, db, user1.ID); err != nil {
			panic(err)
		} else {
			slog.Info(fmt.Sprintf("updated user: %+v", user))
		}

		// delete
		if _, err := user1.Delete(ctx, db); err != nil {
			panic(err)
		}

		if _, err := sqlboiler.FindUser(ctx, db, user1.ID); err != sql.ErrNoRows {
			panic(err)
		}
	}
}

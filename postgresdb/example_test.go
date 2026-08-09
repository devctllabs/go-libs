package postgresdb_test

import (
	"context"
	"log"

	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type user struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func ExampleOpen() {
	ctx := context.Background()
	db, err := postgresdb.Open(ctx, postgresdb.Config{
		Writer: postgresdb.EndpointConfig{DSN: "postgres://app:password@postgres/app"},
		Reader: &postgresdb.EndpointConfig{DSN: "postgres://app:password@postgres-replica/app"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	var users []user
	if err := pgxscan.Select(ctx, db.Reader(), &users, `SELECT id, name FROM users`); err != nil {
		log.Print(err)
		return
	}
}

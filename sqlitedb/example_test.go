package sqlitedb_test

import (
	"context"
	"log"
	"os"

	"github.com/devctllabs/go-libs/sqlitedb"
	"github.com/georgysavva/scany/v2/sqlscan"
)

func ExampleOpen() {
	logger := log.New(os.Stdout, "", 0)
	ctx := context.Background()
	db, err := sqlitedb.Open(ctx, sqlitedb.Config{
		DSN: "file:sqlitedb-example?mode=memory&cache=shared",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	_, err = db.Writer().ExecContext(ctx, `CREATE TABLE messages (body text NOT NULL)`)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = db.Writer().ExecContext(ctx, `INSERT INTO messages (body) VALUES (?)`, "hello")
	if err != nil {
		log.Print(err)
		return
	}
	var messages []struct {
		Body string `db:"body"`
	}
	if err := sqlscan.Select(ctx, db.Reader(), &messages, `SELECT body FROM messages`); err != nil {
		log.Print(err)
		return
	}
	logger.Println(messages[0].Body)
	// Output: hello
}

package di_test

import (
	"context"
	"log"
	"os"

	"github.com/devctllabs/go-libs/di"
)

type exampleDatabase struct {
	name string
}

type exampleRepository struct {
	database *exampleDatabase
}

func Example() {
	logger := log.New(os.Stdout, "", 0)
	container := di.New()
	if err := di.ProvideResource(container,
		func(di.Resolver) (*exampleDatabase, error) {
			return &exampleDatabase{name: "primary"}, nil
		},
		func(_ context.Context, database *exampleDatabase) error {
			logger.Println("close", database.name)
			return nil
		},
	); err != nil {
		log.Fatal(err)
	}
	if err := di.Provide(container, func(resolver di.Resolver) (*exampleRepository, error) {
		database, err := di.Resolve[*exampleDatabase](resolver)
		if err != nil {
			return nil, err
		}
		return &exampleRepository{database: database}, nil
	}); err != nil {
		log.Fatal(err)
	}

	repository, err := di.Resolve[*exampleRepository](container)
	if err != nil {
		log.Fatal(err)
	}
	logger.Println(repository.database.name)
	if err := container.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Output:
	// primary
	// close primary
}

func ExampleProvideNamed() {
	logger := log.New(os.Stdout, "", 0)
	container := di.New()
	if err := di.ProvideNamedValue(container, "primary", "postgres"); err != nil {
		log.Fatal(err)
	}
	if err := di.ProvideNamedValue(container, "analytics", "clickhouse"); err != nil {
		log.Fatal(err)
	}

	value, err := di.ResolveNamed[string](container, "analytics")
	if err != nil {
		log.Fatal(err)
	}
	logger.Println(value)

	// Output: clickhouse
}

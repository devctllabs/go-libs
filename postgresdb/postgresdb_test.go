package postgresdb_test

import (
	"context"
	"testing"

	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/stretchr/testify/require"
)

func TestOpenRejectsBlankWriterDSN(t *testing.T) {
	t.Parallel()
	db, err := postgresdb.Open(context.Background(), postgresdb.Config{})

	require.Nil(t, db)
	require.Error(t, err)
}

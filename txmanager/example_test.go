package txmanager_test

import (
	"context"

	"github.com/devctllabs/go-libs/txmanager"
)

type accountStore interface {
	Debit(ctx context.Context, accountID string, amount int64) error
	Credit(ctx context.Context, accountID string, amount int64) error
}

func transfer(
	ctx context.Context,
	transactions txmanager.Managers,
	accounts accountStore,
	from string,
	to string,
	amount int64,
) error {
	return transactions.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		if err := accounts.Debit(txCtx, from, amount); err != nil {
			return err
		}
		return accounts.Credit(txCtx, to, amount)
	})
}

func ExampleManagers() {
	// Application services can accept txmanager.Managers regardless of whether
	// the composition root selected SQLite or PostgreSQL.
	_ = transfer
}

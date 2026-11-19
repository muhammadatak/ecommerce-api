package domain

import "context"

type TxRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

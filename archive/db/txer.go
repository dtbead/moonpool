package db

import (
	"context"
	"database/sql"
	"errors"
)

type TX interface {
	Commit() error
	Rollback() error
}

type BeginTxer interface {
	BeginTx(ctx context.Context, opt *sql.TxOptions) (Querier, TX, error)
}

func (q *Queries) BeginTx(ctx context.Context, opt *sql.TxOptions) (Querier, TX, error) {
	txer, ok := q.db.(*sql.DB)
	if !ok {
		return nil, nil, errors.New("db is not an sql.DB")
	}

	tx, err := txer.BeginTx(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	return q.WithTx(tx), tx, nil
}

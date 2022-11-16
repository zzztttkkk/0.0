package sqlx

import (
	"context"
	"database/sql"
	"fmt"
)

type Tx struct {
	std       *sql.Tx
	db        *DB
	savepoint string
	ctx       context.Context
	readonly  bool
	canceled  bool
}

func (tx *Tx) Driver() Driver {
	return tx.db.driver
}

func (tx *Tx) DB() *DB {
	return tx.db
}

func (tx *Tx) Raw() *sql.Tx { return tx.std }

func (tx *Tx) BindParams(query string, params interface{}) (string, []interface{}, error) {
	return tx.db.BindParams(query, params)
}

func (tx *Tx) Execute(ctx context.Context, query string, params interface{}) (sql.Result, error) {
	q, a, e := tx.BindParams(query, params)
	if e != nil {
		return nil, e
	}
	return tx.std.ExecContext(ctx, q, a...)
}

func (tx *Tx) Rows(ctx context.Context, query string, params interface{}) (*Rows, error) {
	q, a, e := tx.BindParams(query, params)
	if e != nil {
		return nil, e
	}
	rows, err := tx.std.QueryContext(ctx, q, a...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: rows}, nil
}

func (tx *Tx) FetchOne(ctx context.Context, query string, params interface{}, dist interface{}) error {
	return fetchOne(ctx, tx, query, params, dist)
}

func (tx *Tx) FetchMany(ctx context.Context, query string, params interface{}, slicePtr interface{}) error {
	return fetchMany(ctx, tx, query, params, slicePtr)
}

func (tx *Tx) FetchOneJoined(ctx context.Context, query string, params interface{}, dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	return getJoined(ctx, tx, query, params, dist, joinedGet)
}

func (tx *Tx) FetchManyJoined(ctx context.Context, query string, params interface{}, dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	return selectJoined(ctx, tx, query, params, dist, joinedGet)
}

func (tx *Tx) Prepare(ctx context.Context, query string) (*Stmt, error) {
	query, keys := ScanParams(query, tx.db.driver)
	stmt, err := tx.std.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: stmt prepared by tx: %s, Stmt(%p), Tx(%p)", query, stmt, tx.std)
	}
	return &Stmt{
		std:    stmt,
		keys:   keys,
		logger: tx.db.logger,
	}, nil
}

func (tx *Tx) Stmt(stmt *Stmt) *Stmt {
	v := tx.std.Stmt(stmt.std)
	if tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: tx wrap stmt: (%p)=>(%p), Tx(%p)", stmt, v, tx.std)
	}
	return &Stmt{std: v, keys: stmt.keys, logger: stmt.logger}
}

var _ Executor = (*Tx)(nil)

func (tx *Tx) BeginTx(ctx context.Context, savepoint string) (*Tx, error) {
	if tx.readonly {
		return nil, ErrReadonly
	}
	if tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: tx begin via savepoint, `%s`, Tx(%p);", savepoint, tx.std)
	}

	_, err := tx.Execute(ctx, fmt.Sprintf("SAVEPOINT %s_BEGIN", savepoint), nil)
	if err != nil {
		return nil, err
	}
	return &Tx{std: tx.std, db: tx.db, savepoint: savepoint, ctx: ctx}, nil
}

func (tx *Tx) MustBeginTx(ctx context.Context, savepoint string) *Tx {
	t, e := tx.BeginTx(ctx, savepoint)
	if e != nil {
		panic(e)
	}
	return t
}

func (tx *Tx) Commit() error {
	if tx.canceled {
		return nil
	}

	if len(tx.savepoint) < 1 {
		if tx.db.logger != nil {
			tx.db.logger.Printf("0.0/internal/sqlx: tx commit, Tx(%p);", tx.std)
		}
		return tx.std.Commit()
	}
	if tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: tx commit via savepoint, `%s`, Tx(%p);", tx.savepoint, tx.std)
	}
	_, err := tx.Execute(tx.ctx, fmt.Sprintf("SAVEPOINT %s", tx.savepoint), nil)
	if err != nil {
		return err
	}
	_, err = tx.Execute(tx.ctx, fmt.Sprintf("RELASE SAVEPOINT %s_BEGIN", tx.savepoint), nil)
	return err
}

func (tx *Tx) Rollback() error {
	if len(tx.savepoint) < 1 {
		if tx.db.logger != nil {
			tx.db.logger.Printf("0.0/internal/sqlx: tx rollback, Tx(%p);", tx.std)
		}
		return tx.std.Rollback()
	}
	if tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: tx rollback via savepoint, `%s`, Tx(%p);", tx.savepoint, tx.std)
	}
	_, err := tx.Execute(tx.ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s_BEGIN", tx.savepoint), nil)
	return err
}

func (tx *Tx) RollbackTo(savepoint string) error {
	_, err := tx.Execute(tx.ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepoint), nil)
	return err
}

func (tx *Tx) AutoCommit() {
	var e error
	var v interface{}
	if !tx.canceled {
		v = recover()
		if v == nil {
			if e = tx.Commit(); e == nil {
				return
			}
			if tx.db.logger != nil {
				tx.db.logger.Printf("0.0/internal/sqlx: commit error, %s", e)
			}
		}
	}
	e = tx.Rollback()
	if e != nil && tx.db.logger != nil {
		tx.db.logger.Printf("0.0/internal/sqlx: rollback error: %s", e)
	}
	if v != nil {
		panic(v)
	}
}

func (tx *Tx) Cancel() { tx.canceled = true }

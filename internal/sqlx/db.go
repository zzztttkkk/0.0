package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type DB struct {
	readonly bool
	std      *sql.DB
	logger   Logger
	driver   Driver
}

func OpenDB(driver Driver, dsn string, readonly bool, logger Logger) (*DB, error) {
	connector, e := driver.Open(dsn)
	if e != nil {
		return nil, e
	}

	db := &DB{
		readonly: readonly,
		std:      sql.OpenDB(connector),
		logger:   logger,
		driver:   driver,
	}
	return db, nil
}

func MustOpenDB(driver Driver, dsn string, readonly bool, logger Logger) *DB {
	if !strings.HasPrefix(dsn, "postgres://") {
		dsn = "postgres://" + dsn
	}
	db, err := OpenDB(driver, dsn, readonly, logger)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *DB) Driver() Driver {
	return db.driver
}

// DB for interface `Executor`
func (db *DB) DB() *DB {
	return db
}

func (db *DB) Raw() *sql.DB { return db.std }

func (db *DB) BindParams(query string, params interface{}) (string, []interface{}, error) {
	q, keys := ScanParams(query, db.driver)
	if len(keys) < 1 {
		return query, nil, nil
	}

	args, err := paramsToArgs(params, keys)
	if err != nil {
		return "", nil, err
	}

	if db.logger != nil {
		db.logger.Printf("%s %v", q, args)
	}
	return q, args, nil
}

func (db *DB) Execute(ctx context.Context, query string, params interface{}) (sql.Result, error) {
	q, a, e := db.BindParams(query, params)
	if e != nil {
		return nil, e
	}
	return db.std.ExecContext(ctx, q, a...)
}

func (db *DB) Rows(ctx context.Context, query string, params interface{}) (*Rows, error) {
	q, a, e := db.BindParams(query, params)
	if e != nil {
		return nil, e
	}
	rows, err := db.std.QueryContext(ctx, q, a...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: rows}, nil
}

func (db *DB) FetchOne(ctx context.Context, query string, params interface{}, dist interface{}) error {
	return fetchOne(ctx, db, query, params, dist)
}

func (db *DB) FetchMany(ctx context.Context, query string, params interface{}, slicePtr interface{}) error {
	return fetchMany(ctx, db, query, params, slicePtr)
}

func (db *DB) FetchOneJoined(ctx context.Context, query string, params interface{}, dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	return getJoined(ctx, db, query, params, dist, joinedGet)
}

func (db *DB) FetchManyJoined(ctx context.Context, query string, params interface{}, dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	return selectJoined(ctx, db, query, params, dist, joinedGet)
}

var ErrReadonly = errors.New("0.0/internal/sqlx: readonly tx")

func (db *DB) BeginTx(ctx context.Context, opt *sql.TxOptions) (*Tx, error) {
	var readonly = false
	if opt != nil {
		readonly = opt.ReadOnly
	}
	if db.readonly && !readonly {
		return nil, ErrReadonly
	}

	tx, err := db.std.BeginTx(ctx, opt)
	if err != nil {
		return nil, err
	}
	if db.logger != nil {
		db.logger.Printf("0.0/internal/sqlx: tx begin, Tx(%p);", tx)
	}
	return &Tx{std: tx, db: db, ctx: ctx, readonly: readonly}, nil
}

func (db *DB) MustBeginTx(ctx context.Context, opt *sql.TxOptions) *Tx {
	t, e := db.BeginTx(ctx, opt)
	if e != nil {
		panic(e)
	}
	return t
}

func (db *DB) Prepare(ctx context.Context, query string) (*Stmt, error) {
	query, keys := ScanParams(query, db.driver)
	stmt, err := db.std.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if db.logger != nil {
		db.logger.Printf("0.0/internal/sqlx: stmt prepared by db: %s, Stmt(%p)", query, stmt)
	}
	return &Stmt{
		std:    stmt,
		keys:   keys,
		logger: db.logger,
	}, nil
}

var _ Executor = (*DB)(nil)

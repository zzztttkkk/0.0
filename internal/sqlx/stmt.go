package sqlx

import (
	"context"
	"database/sql"
)

type Stmt struct {
	std    *sql.Stmt
	keys   []string
	logger Logger
}

func (stmt *Stmt) Close() error {
	if stmt.logger != nil {
		stmt.logger.Printf("stmt close, tsql.Stmt(%p)", stmt.std)
	}
	return stmt.std.Close()
}

func (stmt *Stmt) Execute(ctx context.Context, params interface{}) (sql.Result, error) {
	args, err := paramsToArgs(params, stmt.keys)
	if err != nil {
		return nil, err
	}
	if stmt.logger != nil {
		stmt.logger.Printf("stmt execute, args(%v), tsql.Stmt(%p)", args, stmt.std)
	}
	return stmt.std.ExecContext(ctx, args...)
}

func (stmt *Stmt) Rows(ctx context.Context, params interface{}) (*Rows, error) {
	args, err := paramsToArgs(params, stmt.keys)
	if err != nil {
		return nil, err
	}
	if stmt.logger != nil {
		stmt.logger.Printf("stmt select, args(%v), tsql.Stmt(%p)", args, stmt.std)
	}
	rows, err := stmt.std.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: rows}, nil
}

func (stmt *Stmt) FetchOne(ctx context.Context, params interface{}, dist interface{}) error {
	rows, err := stmt.Rows(ctx, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.fetchOne(dist)
}

func (stmt *Stmt) FetchMany(ctx context.Context, params interface{}, dist interface{}) error {
	rows, err := stmt.Rows(ctx, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.fetchMany(dist)
}

func (stmt *Stmt) FetchOneJoined(ctx context.Context, params interface{}, dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	rows, err := stmt.Rows(ctx, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.getJoined(dist, joinedGet)
}

func (stmt *Stmt) FetchManyJoined(ctx context.Context, params interface{}, ptrOfJoinedDistSlice interface{}, joinedGet JoinedEmbedDistGetter) error {
	rows, err := stmt.Rows(ctx, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.selectJoined(ptrOfJoinedDistSlice, joinedGet)
}

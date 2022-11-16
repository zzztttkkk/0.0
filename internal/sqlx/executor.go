package sqlx

import (
	"context"
	"database/sql"
)

type BasicExecutor interface {
	Execute(ctx context.Context, query string, params interface{}) (sql.Result, error)
	Rows(ctx context.Context, query string, params interface{}) (*Rows, error)
	FetchOne(ctx context.Context, query string, params interface{}, ptrOfDist interface{}) error
	FetchMany(ctx context.Context, query string, params interface{}, ptrOfDistSlice interface{}) error
	FetchOneJoined(ctx context.Context, query string, params interface{}, dist interface{}, get JoinedEmbedDistGetter) error
	FetchManyJoined(ctx context.Context, query string, params interface{}, ptrOfJoinedDistSlice interface{}, get JoinedEmbedDistGetter) error
}

type Executor interface {
	BasicExecutor
	BindParams(query string, params interface{}) (string, []interface{}, error)
	Prepare(ctx context.Context, query string) (*Stmt, error)
	DB() *DB
	Driver() Driver
}

func fetchOne(ctx context.Context, be BasicExecutor, query string, params interface{}, dist interface{}) error {
	rows, err := be.Rows(ctx, query, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.fetchOne(dist)
}

func fetchMany(ctx context.Context, be BasicExecutor, query string, params interface{}, slicePtr interface{}) error {
	rows, err := be.Rows(ctx, query, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.fetchMany(slicePtr)
}

func getJoined(ctx context.Context, be BasicExecutor, query string, params interface{}, dist interface{}, get JoinedEmbedDistGetter) error {
	rows, err := be.Rows(ctx, query, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.getJoined(dist, get)
}

func selectJoined(ctx context.Context, be BasicExecutor, query string, params interface{}, slicePtr interface{}, joinedGet JoinedEmbedDistGetter) error {
	rows, err := be.Rows(ctx, query, params)
	if err != nil {
		return err
	}
	defer rows.Close()
	return rows.selectJoined(slicePtr, joinedGet)
}

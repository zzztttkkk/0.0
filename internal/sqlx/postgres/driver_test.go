package postgres

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"testing"
)

func TestPostgres(t *testing.T) {
	db := sqlx.MustOpenDB(&Driver{}, "ztk:123456@localhost:5432/local_test", false, nil)

	sum := 0
	err := db.FetchOne(context.Background(), "select 1 + ${a} + ${a}", sqlx.Params{"a": 45}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)
}

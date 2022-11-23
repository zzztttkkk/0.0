package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"testing"
	"time"
)

type Base struct {
	CreatedAt uint64        `db:"created_at;default=(extract(epoch from now()) * 1000)::bigint"`
	DeletedAt sql.NullInt64 `db:"deleted_at;nullable"`
}

type Xyz struct {
	Base
	V1 int64     `db:"v1;primary;incr;unique"`
	V2 time.Time `db:"v2;default=now();index=x1unique|x2unique,asc"`
	V3 *AnyJSON  `db:"v3;sqltype=json;nullable"`
	V4 *uint64   `db:"v4;nullable"`
	V5 string    `db:"v5;length=~30;default='';index=x1unique"`
}

func TestPostgres(t *testing.T) {
	db := Open("postgres:123456@localhost:15432/o_o", false, nil)
	db.EnableHStore(context.Background()).EnableUUID(context.Background())

	sum := 0.0
	err := db.FetchOne(context.Background(), "select 1 + ${a}::float + ${a}::float as sum", sqlx.Params{"a": 4.9}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)

	fmt.Println(db.CreateTable(context.Background(), Xyz{}))

	var xyz Xyz
	err = db.FetchOne(context.Background(), "select * from xyz where v1=${v1}", sqlx.Params{"v1": 2}, &xyz)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\r\n", xyz)
	}
}

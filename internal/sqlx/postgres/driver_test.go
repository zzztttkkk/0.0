package postgres

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zzztttkkk/0.0/internal/sqlx"
)

type Base struct {
	CreatedAt uint64 `db:"created_at"`
	DeletedAt uint64 `db:"deleted_at"`
}

type User struct {
	Base
	Id       uint64        `db:"id,ddl=numeric(19) check (id >= 0)"`
	Uuid     pgtype.UUID   `db:"uuid,default=uuid"`
	Name     string        `db:"name"`
	Nickname string        `db:"nickname"`
	Ext      pgtype.Hstore `db:"ext"`
}

func TestPostgres(t *testing.T) {
	db := sqlx.MustOpenDB(&Driver{}, "ztk:123456@localhost:5432/local_test", false, nil)

	sum := 0.0
	err := db.FetchOne(context.Background(), "select 1 + ${a}::float + ${a}::float as sum", sqlx.Params{"a": 4.9}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)

	var num uint8 = 255
	fmt.Println(num, math.MaxUint32)

	db.Driver().DDL(User{})
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"testing"
	"time"
)

type Base struct {
	CreatedAt uint64 `db:"created_at,default=(extract(epoch from now()) * 1000)::bigint"`
	DeletedAt uint64 `db:"deleted_at"`
}

type User struct {
	Base
	Id       uint64         `db:"id,primary"`
	Uuid     pgtype.UUID    `db:"uuid,default=uuid_generate_v4(),unique"`
	Name     string         `db:"name,length=~30,unique"`
	Nickname sql.NullString `db:"nickname,length=~30"`
	Ext      pgtype.Hstore  `db:"ext"`
}

type Xyz struct {
	V1 int64     `db:"v1"`
	V2 time.Time `db:"v2"`
	V3 AnyJSON   `db:"v3"`
	V4 int64     `db:"v4"`
}

func TestPostgres(t *testing.T) {
	db := Open("postgres:123456@localhost:5432/local_test", false, nil)
	db.EnableHStore(context.Background()).EnableUUID(context.Background())

	sum := 0.0
	err := db.FetchOne(context.Background(), "select 1 + ${a}::float + ${a}::float as sum", sqlx.Params{"a": 4.9}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)

	var xyz Xyz
	err = db.FetchOne(context.Background(), "select * from xyz where v1=${v1}", sqlx.Params{"v1": 2}, &xyz)
	fmt.Println(err, xyz)

	db.CreateTable(User{})
}

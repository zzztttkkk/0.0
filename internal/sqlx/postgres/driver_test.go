package postgres

import (
	"context"
	"database/sql"
	"fmt"
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
	Id       uint64         `db:"id,primary"`
	Uuid     pgtype.UUID    `db:"uuid,default=uuid_generate_v4(),unique"`
	Name     string         `db:"name,length=~30,unique"`
	Nickname sql.NullString `db:"nickname,length=~30"`
	Ext      pgtype.Hstore  `db:"ext"`
}

func (_ User) DDLId() *sqlx.FieldDefinition {
	return nil
}

func TestPostgres(t *testing.T) {
	db := Open("ztk:123456@localhost:5432/local_test", false, nil)
	db.EnableHStore(context.Background()).EnableUUID(context.Background())

	sum := 0.0
	err := db.FetchOne(context.Background(), "select 1 + ${a}::float + ${a}::float as sum", sqlx.Params{"a": 4.9}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)

	db.CreateTable(User{})
}

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zzztttkkk/0.0/internal/sqlx"
)

type User struct {
	Id       int64       `db:"id"`
	Uuid     pgtype.UUID `db:"uuid"`
	Name     string      `db:"name"`
	Nickname string      `db:"nickname"`
}

func TestPostgres(t *testing.T) {
	db := sqlx.MustOpenDB(&Driver{}, "postgres:123456@localhost:5432/local_test", false, nil)

	sum := 0.0
	err := db.FetchOne(context.Background(), "select 1 + ${a}::float + ${a}::float as sum", sqlx.Params{"a": 4.9}, &sum)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sum)

	var user User
	err = db.FetchOne(context.Background(), "select * from public.\"User\" where nickname=${nickname}", sqlx.Params{"nickname": "😄"}, &user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v, %s", user, "😄")
}

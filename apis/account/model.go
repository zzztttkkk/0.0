package account

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zzztttkkk/0.0/apis/common"
	"github.com/zzztttkkk/0.0/config"
	"github.com/zzztttkkk/0.0/internal"
)

type DBAccountUser struct {
	common.BaseModel
	Id         int64          `db:"id;incr;primary;unique"`
	Uuid       pgtype.UUID    `db:"uuid;unique;default=uuid_generate_v4()"`
	Email      string         `db:"email;length=~120;unique"`
	Nickname   string         `db:"nickname;length=~30"`
	Avatar     *string        `db:"avatar;length=~120;nullable"`
	Bio        *string        `db:"bio;length=~245;nullable"`
	ExtPubInfo *pgtype.Hstore `db:"extpubinfo;nullable"`
}

func init() {
	internal.LazyInvoke(func(cfg *config.Config) {
		db := cfg.DBMaster()
		if err := db.CreateTable(cfg.Context(), DBAccountUser{}); err != nil {
			panic(err)
		}
	})
}

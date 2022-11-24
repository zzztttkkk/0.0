package common

type BaseModel struct {
	CreatedAt int64 `db:"created_at,default=(extract(epoch from now()) * 1000)::bigint"`
	DeletedAt int64 `db:"deleted_at,nullable"`
}

package sqlx

import (
	"context"
	"database/sql"
	"github.com/zzztttkkk/0.0/internal/sqlx/internal"
	"math/rand"
	"time"
)

type _CtxKey int

const (
	_KeyDB = _CtxKey(iota + 1)
	_KeyJustWDB
	_KeyTx
)

type Logger interface {
	Printf(fmt string, args ...interface{})
}

type GroupOptions struct {
	ReadonlySourceNames []string
	MaxIdleConns        int
	MaxOpenConns        int
	ConnMaxLifetime     int
	ConnMaxIdleTime     int
	Logger              Logger
}

type Group struct {
	opts   GroupOptions
	w      *DB
	rs     []*DB
	pick   func() int
	driver internal.Driver
	logger Logger
}

func (g *Group) Driver() internal.Driver {
	return g.driver
}

func (g *Group) Execute(ctx context.Context, query string, params interface{}) (sql.Result, error) {
	return g.w.Execute(ctx, query, params)
}

func (g *Group) Rows(ctx context.Context, query string, params interface{}) (*Rows, error) {
	return g.w.Rows(ctx, query, params)
}

func (g *Group) FetchOne(ctx context.Context, query string, params interface{}, dist interface{}) error {
	return g.w.FetchOne(ctx, query, params, dist)
}

func (g *Group) FetchMany(ctx context.Context, query string, params interface{}, ptrOfDistSlice interface{}) error {
	return g.w.FetchMany(ctx, query, params, ptrOfDistSlice)
}

func (g *Group) FetchOneJoined(ctx context.Context, query string, params interface{}, dist interface{}, get JoinedEmbedDistGetter) error {
	return g.w.FetchOneJoined(ctx, query, params, dist, get)
}

func (g *Group) FetchManyJoined(ctx context.Context, query string, params interface{}, ptrOfJoinedDistSlice interface{}, get JoinedEmbedDistGetter) error {
	return g.w.FetchManyJoined(ctx, query, params, ptrOfJoinedDistSlice, get)
}

func (g *Group) BindParams(query string, params interface{}) (string, []interface{}, error) {
	return g.w.BindParams(query, params)
}

func (g *Group) Prepare(ctx context.Context, query string) (*Stmt, error) {
	return g.w.Prepare(ctx, query)
}

func (g *Group) DB() *DB {
	return g.w
}

var (
	_ Executor = (*Group)(nil)
)

func NewGroup(driver internal.Driver, dsn string, opts *GroupOptions) *Group {
	if opts == nil {
		opts = &GroupOptions{}
	}
	g := &Group{driver: driver}
	g.opts = *opts
	g.logger = g.opts.Logger
	g.w = g.mustOpen(dsn)
	for _, dsn := range g.opts.ReadonlySourceNames {
		v := g.mustOpen(dsn)
		v.readonly = true
		g.rs = append(g.rs, v)
	}
	return g
}

func (g *Group) SetLogger(logger Logger) *Group {
	g.opts.Logger = logger
	g.logger = logger
	return g
}

func (g *Group) open(dsn string) (*DB, error) {
	c, e := g.driver.Open(dsn)
	if e != nil {
		return nil, e
	}
	stdDB := sql.OpenDB(c)
	if g.opts.MaxOpenConns > 0 {
		stdDB.SetMaxOpenConns(g.opts.MaxOpenConns)
	}
	if g.opts.MaxIdleConns > 0 {
		stdDB.SetMaxIdleConns(g.opts.MaxIdleConns)
	}
	if g.opts.ConnMaxIdleTime > 0 {
		stdDB.SetConnMaxIdleTime(time.Second * time.Duration(g.opts.ConnMaxIdleTime))
	}
	if g.opts.ConnMaxLifetime > 0 {
		stdDB.SetConnMaxLifetime(time.Second * time.Duration(g.opts.ConnMaxLifetime))
	}
	if g.logger != nil {
		g.logger.Printf("0.0/internal/sqlx: open database %s", dsn)
	}
	return &DB{std: stdDB, logger: g.logger, driver: g.driver}, nil
}

func (g *Group) mustOpen(dsn string) *DB {
	v, e := g.open(dsn)
	if e != nil {
		panic(e)
	}
	return v
}

func (g *Group) SetPick(fn func() int) *Group {
	g.pick = fn
	return g
}

func (g *Group) pickRDB() *DB {
	if len(g.rs) < 1 {
		return g.w
	}
	idx := 0
	if g.pick != nil {
		idx = g.pick() % len(g.rs)
	} else {
		idx = rand.Int() % len(g.rs)
	}
	return g.rs[idx]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (g *Group) PickExecutor(ctx context.Context) (context.Context, Executor) {
	exe := getExe(ctx)
	if exe != nil {
		return ctx, exe
	}

	var db *DB
	if ctx.Value(_KeyJustWDB) != nil {
		db = g.w
	} else {
		db = g.pickRDB()
	}
	return context.WithValue(ctx, _KeyDB, db), db
}

func getExe(ctx context.Context) Executor {
	txi := ctx.Value(_KeyTx)
	if txi != nil {
		return txi.(*Tx)
	}
	dbi := ctx.Value(_KeyDB)
	if dbi != nil {
		return dbi.(*DB)
	}
	return nil
}

type TxOptions struct {
	sql.TxOptions
	Savepoint      string
	JustWritableDB bool
}

func (g *Group) MustBegin(ctx context.Context, opts *TxOptions) (context.Context, *Tx) {
	exe := getExe(ctx)
	var rTx *Tx
	if exe != nil {
		tx, ok := exe.(*Tx)
		if ok { // options can not be nil
			rTx = tx.MustBeginTx(ctx, opts.Savepoint)
		}
	}
	if rTx == nil { // options can be nil
		if opts != nil {
			var db = g.w
			if opts.ReadOnly && !opts.JustWritableDB {
				db = g.pickRDB()
			}
			rTx = db.MustBeginTx(ctx, &opts.TxOptions)
		} else {
			rTx = g.w.MustBeginTx(ctx, nil)
		}
	}
	return context.WithValue(ctx, _KeyTx, rTx), rTx
}

func (g *Group) ReadonlyDB() *DB {
	return g.pickRDB()
}

func (g *Group) JustWritableDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, _KeyJustWDB, true)
}

//func (g *Group) CreateTable(ctx context.Context, model *Model) error {
//	return g.w.CreateTable(ctx, model)
//}

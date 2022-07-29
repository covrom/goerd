package goerd_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/covrom/goerd"
	"github.com/covrom/goerd/schema"
	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime

	CategoryID uuid.UUID
	Category   Category

	Name string
	Code string
	Unit string
}

type Products struct {
	sc      *schema.Schema
	t       *schema.Table
	cols    []string
	fields  func(Product) []interface{}
	pfields func(*Product) []interface{}
}

func NewProducts(sc *schema.Schema) (*Products, error) {
	tname := "products"

	t, err := sc.FindTableByName(tname)
	if err != nil {
		return nil, err
	}

	p := &Products{
		sc: sc,
		t:  t,

		cols: []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"name",
			"category_id",
			"code",
			"unit",
		},

		fields: func(p Product) []interface{} {
			return []interface{}{
				p.ID,
				p.CreatedAt,
				p.UpdatedAt,
				p.DeletedAt,
				p.Name,
				p.CategoryID,
				p.Code,
				p.Unit,
			}
		},

		pfields: func(p *Product) []interface{} {
			return []interface{}{
				&p.ID,
				&p.CreatedAt,
				&p.UpdatedAt,
				&p.DeletedAt,
				&p.Name,
				&p.CategoryID,
				&p.Code,
				&p.Unit,
			}
		},
	}

	for _, c := range p.cols {
		if _, err := t.FindColumnByName(c); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (d *Products) Table() string {
	return d.t.Name
}

func (d *Products) Columns() []string {
	return d.cols
}

// nolint hugeParam
func (d *Products) Fields(p Product) []interface{} {
	return d.fields(p)
}

func (d *Products) PFields(p *Product) []interface{} {
	return d.pfields(p)
}

// nolint hugeParam
func (d *Products) ProductToStore(ctx context.Context, p Product) error {
	return goerd.WithTx(ctx, func(ctxTx context.Context) error {
		q := goerd.ReplaceQuery(d, "id")
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		_, err := goerd.SqlxTxFromContext(ctxTx).
			ExecContext(ctxTx, q, d.Fields(p)...)
		return err
	})
}

type Identity struct {
	ID        uuid.UUID    `db:"id"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (d *Products) AllProductIDs(ctx context.Context) ([]Identity, error) {
	var dbidts []Identity

	err := goerd.WithTx(ctx, func(ctxTx context.Context) error {
		return goerd.SqlxTxFromContext(ctxTx).
			SelectContext(ctx, &dbidts,
				fmt.Sprintf(`select id, updated_at, deleted_at from %s`,
					d.Table()))
	})
	if err != nil {
		return nil, err
	}

	return dbidts, nil
}

package goerd_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/covrom/goerd"
	"github.com/covrom/goerd/schema"
	"github.com/google/uuid"
)

type Category struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
	ParentID   *uuid.UUID
	Name       string
	IsDisabled bool
}

type Categorys struct {
	sc      *schema.Schema
	t       *schema.Table
	cols    []string
	fields  func(Category) []interface{}
	pfields func(*Category) []interface{}
}

func NewCategorys(sc *schema.Schema) (*Categorys, error) {
	tname := "categories"
	t, err := sc.FindTableByName(tname)
	if err != nil {
		return nil, err
	}
	cs := &Categorys{
		sc: sc,
		t:  t,

		cols: []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"parent_id",
			"name",
			"is_disabled",
		},

		fields: func(p Category) []interface{} {
			return []interface{}{
				p.ID,
				p.CreatedAt,
				p.UpdatedAt,
				p.DeletedAt,
				p.ParentID,
				p.Name,
				p.IsDisabled,
			}
		},

		pfields: func(p *Category) []interface{} {
			return []interface{}{
				&p.ID,
				&p.CreatedAt,
				&p.UpdatedAt,
				&p.DeletedAt,
				&p.ParentID,
				&p.Name,
				&p.IsDisabled,
			}
		},
	}

	for _, c := range cs.cols {
		if _, err := t.FindColumnByName(c); err != nil {
			return nil, err
		}
	}
	return cs, nil
}

func (d *Categorys) Table() string {
	return d.t.Name
}

func (d *Categorys) Columns() []string {
	return d.cols
}

// nolint hugeParam
func (d *Categorys) Fields(p Category) []interface{} {
	return d.fields(p)
}

func (d *Categorys) PFields(p *Category) []interface{} {
	return d.pfields(p)
}

func (d *Categorys) ScanFields(p *Category) []interface{} {
	return []interface{}{
		&p.ID,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.DeletedAt,
		&p.ParentID,
		&p.Name,
		&p.IsDisabled,
	}
}

// nolint hugeParam
func (d *Categorys) CategoryToStore(ctx context.Context, p Category) error {
	return goerd.WithTx(ctx, func(ctxTx context.Context) error {
		q := goerd.ReplaceQuery(d, "id")
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		_, err := goerd.SqlxTxFromContext(ctxTx).
			ExecContext(ctx, q, d.Fields(p)...)
		return err
	})
}

func (d *Categorys) ListCategoriesUpdatedFrom(ctx context.Context, updatedFrom time.Time) ([]Category, error) {
	ret := make([]Category, 0, 100)

	err := goerd.WithTx(ctx, func(ctxTx context.Context) error {
		q := `select %s from %s where $1 OR updated_at >= $2`

		rows, err := goerd.SqlxTxFromContext(ctxTx).
			QueryxContext(ctx, fmt.Sprintf(q,
				strings.Join(d.Columns(), ","),
				d.Table()), updatedFrom.IsZero(), updatedFrom)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			p := &Category{}
			scanFields := d.ScanFields(p)
			if err := rows.Scan(scanFields...); err != nil {
				return err
			}
			ret = append(ret, *p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

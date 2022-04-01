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
	ParentID   uuid.UUID
	Name       string
	IsDisabled bool
}

type Categorys struct{}

func NewCategorys() *Categorys {
	return &Categorys{}
}

func (d *Categorys) Table() string {
	return "categories"
}

func (d *Categorys) Columns() []string {
	return []string{
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"parent_id",
		"name",
		"is_disabled",
	}
}

// nolint hugeParam
func (d *Categorys) Fields(p Category) []interface{} {
	return []interface{}{
		p.ID,
		p.CreatedAt,
		p.UpdatedAt,
		p.DeletedAt,
		p.ParentID,
		p.Name,
		p.IsDisabled,
	}
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

func (d *Categorys) TableDef() *schema.Table {
	return &schema.Table{
		Name: d.Table(),
		Columns: []*schema.Column{
			{
				Name:       "id",
				Type:       "uuid",
				PrimaryKey: true,
			},
			{
				Name: "created_at",
				Type: "timestamptz",
			},
			{
				Name: "updated_at",
				Type: "timestamptz",
			},
			{
				Name:     "deleted_at",
				Type:     "timestamptz",
				Nullable: true,
			},
			{
				Name: "parent_id",
				Type: "uuid",
			},
			{
				Name: "name",
				Type: "varchar(200)",
			},
			{
				Name: "is_disabled",
				Type: "boolean",
				Default: sql.NullString{
					String: "false",
					Valid:  true,
				},
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "categories_deleted_at",
				Columns: []string{"deleted_at"},
			},
			{
				Name:    "categories_parent_id",
				Columns: []string{"parent_id"},
			},
		},
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

func CategoryModel() *goerd.ObjectModel[Category] {
	md := goerd.Model[Category](
		"categories",
		goerd.Field[uuid.UUID](&schema.Column{
			Name:       "id",
			Type:       "uuid",
			PrimaryKey: true,
		}),
		goerd.Field[time.Time](&schema.Column{
			Name: "created_at",
			Type: "timestamptz",
		}),
		goerd.Field[time.Time](&schema.Column{
			Name: "updated_at",
			Type: "timestamptz",
		}),
		goerd.Field[sql.NullTime](&schema.Column{
			Name:     "deleted_at",
			Type:     "timestamptz",
			Nullable: true,
		}),
		goerd.Field[uuid.UUID](&schema.Column{
			Name: "parent_id",
			Type: "uuid",
		}),
		goerd.Field[string](&schema.Column{
			Name: "name",
			Type: "varchar(200)",
		}),
		goerd.Field[bool](&schema.Column{
			Name: "is_disabled",
			Type: "boolean",
			Default: sql.NullString{
				String: "false",
				Valid:  true,
			},
		}),
	)
	return md
}

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

type Products struct{}

func NewProducts() *Products {
	return &Products{}
}

func (d *Products) Table() string {
	return "products"
}

func (d *Products) Columns() []string {
	return []string{
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"name",
		"category_id",
		"code",
		"unit",
	}
}

// nolint hugeParam
func (d *Products) Fields(p Product) []interface{} {
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
}

func (d *Products) TableDef() *schema.Table {
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
				Name: "category_id",
				Type: "uuid",
			},
			{
				Name: "name",
				Type: "varchar(200)",
			},
			{
				Name: "code",
				Type: "varchar(80)",
			},
			{
				Name: "unit",
				Type: "varchar(30)",
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "products_deleted_at",
				Columns: []string{"deleted_at"},
			},
			{
				Name:    "products_category_id",
				Columns: []string{"category_id"},
			},
			{
				Name:    "products_code",
				Columns: []string{"code"},
			},
		},
	}
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

func ProductModel() *goerd.ObjectModel[Product] {
	md := goerd.Model[Product](
		"products",
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
			Name: "category_id",
			Type: "uuid",
		}),
		goerd.Field[string](&schema.Column{
			Name: "name",
			Type: "varchar(200)",
		}),
		goerd.Field[string](&schema.Column{
			Name: "code",
			Type: "varchar(80)",
		}),
		goerd.Field[string](&schema.Column{
			Name: "unit",
			Type: "varchar(30)",
		}),
	).WithIndex(
		&schema.Index{
			Name:    "products_deleted_at",
			Columns: []string{"deleted_at"},
		},
		&schema.Index{
			Name:    "products_category_id",
			Columns: []string{"category_id"},
		},
		&schema.Index{
			Name:    "products_code",
			Columns: []string{"code"},
		},
	)
	return md
}

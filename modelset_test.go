package goerd_test

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/covrom/goerd"
	"github.com/covrom/goerd/schema"
	"github.com/google/uuid"
)

func TestModelSet(t *testing.T) {
	if db == nil {
		log.Fatal("run TestMain before")
	}

	mset := ModelTestSet()

	err := mset.Migrate(db, "modeltest")
	if err != nil {
		t.Errorf("Migrate error: %s", err)
		return
	}
}

func ModelTestSet() *goerd.ModelSet {
	cm := CategoryModel()
	cid := cm.Field("id").(*goerd.ObjectField[uuid.UUID])

	pm := ProductModel()
	pcid := pm.Field("category_id").(*goerd.ObjectField[uuid.UUID])

	mset := goerd.NewModelSet(
		cm,
		pm,
	).WithRelations(
		&schema.Relation{
			Name:  "product_category_rel",
			Table: pm.SchemaTable(),
			Columns: []*schema.Column{
				pcid.Column(),
			},
			ParentTable: cm.SchemaTable(),
			ParentColumns: []*schema.Column{
				cid.Column(),
			},
			OnDelete: "CASCADE",
		},
	)
	return mset
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
	).WithIndex(
		&schema.Index{
			Name:    "categories_deleted_at",
			Columns: []string{"deleted_at"},
		},
		&schema.Index{
			Name:    "categories_parent_id",
			Columns: []string{"parent_id"},
		},
	)
	return md
}

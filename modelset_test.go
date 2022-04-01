package goerd_test

import (
	"log"
	"testing"

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

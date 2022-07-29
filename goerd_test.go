package goerd_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/covrom/goerd"
	"github.com/covrom/goerd/schema"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//go:embed shop.yaml
var shopSchemeYaml []byte

func TestBasicUsage(t *testing.T) {
	if db == nil {
		t.Fatal("run TestMain before")
		return
	}

	targetSchema := &schema.Schema{}
	if err := targetSchema.UnmarshalYAML(shopSchemeYaml); err != nil {
		t.Error("targetSchema.UnmarshalYAML error:", err)
		return
	}

	cats, err := NewCategorys(targetSchema)
	if err != nil {
		t.Error("NewCategorys error:", err)
		return
	}
	prods, err := NewProducts(targetSchema)
	if err != nil {
		t.Error("NewProducts error:", err)
		return
	}

	err = Migrate(db, targetSchema)
	if err != nil {
		t.Error("Migrate error:", err)
		return
	}

	c := Category{
		ID:   uuid.New(),
		Name: "category 1",
	}

	p := Product{
		ID:         uuid.New(),
		CategoryID: c.ID,
		Name:       "product 1",
		Code:       "1000",
		Unit:       "pack",
	}

	ctx := goerd.WithSqlxDb(context.Background(), db)

	if err := cats.CategoryToStore(ctx, c); err != nil {
		t.Errorf("CategoryToStore error: %s", err)
		return
	}

	if err := prods.ProductToStore(ctx, p); err != nil {
		t.Errorf("ProductToStore error: %s", err)
		return
	}

	ls, err := cats.ListCategoriesUpdatedFrom(ctx, time.Now().AddDate(-1, 0, 0))
	if err != nil {
		t.Errorf("cats.ListCategoriesUpdatedFrom error: %s", err)
		return
	}

	if len(ls) != 1 {
		t.Errorf("cats.ListCategoriesUpdatedFrom count != 1")
		return
	}

	pids, err := prods.AllProductIDs(ctx)
	if err != nil {
		t.Errorf("prods.AllProductIDs error: %s", err)
		return
	}

	if len(pids) != 1 {
		t.Errorf("prods.AllProductIDs count != 1")
		return
	}
}

func Migrate(d *sqlx.DB, migsch *schema.Schema) error {
	dbsch, err := goerd.SchemaFromPostgresDB(d.DB)
	if err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	qs, err := goerd.GenerateMigrationSQL(dbsch, migsch)
	if err != nil {
		return err
	}
	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	for i, q := range qs {
		// skip all dropping DDL queries
		if strings.HasPrefix(strings.ToUpper(q), "DROP") {
			fmt.Println(i+1, "skip: ", q)
			continue
		}
		if strings.Contains(strings.ToUpper(q), "DROP COLUMN") {
			fmt.Println(i+1, "skip: ", q)
			continue
		}

		fmt.Println(i+1, q)

		_, err = tx.Exec(q)

		if err != nil {
			_ = tx.Rollback()

			fmt.Println("db schema:")
			dbsch.SaveYaml(os.Stdout)
			fmt.Println("target schema:")
			migsch.SaveYaml(os.Stdout)

			return fmt.Errorf("cannot migrate database %q: %w", q, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	return nil
}

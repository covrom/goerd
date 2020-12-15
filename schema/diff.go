package schema

import (
	"strings"

	"github.com/covrom/diff"
)

func (t1 *Table) EqualTo(t2 *Table) bool {
	if t1 == t2 {
		return true
	}

	if t1.Name != t2.Name ||
		t1.Type != t2.Type ||
		t1.Def != t2.Def {
		return false
	}

	// compare columns
	ct := ColumnTransition{
		From: t1.Columns,
		To:   t2.Columns,
	}
	ch := diff.Diff(len(ct.From), len(ct.To), &ct)
	if len(ch) != 0 {
		return false
	}

	// compare indexes
	cit := IndexTransition{
		From: t1.Indexes,
		To:   t2.Indexes,
	}
	ch = diff.Diff(len(cit.From), len(cit.To), &cit)
	if len(ch) != 0 {
		return false
	}

	// compare constraints
	ctt := ConstraintTransition{
		From: t1.Constraints,
		To:   t2.Constraints,
	}
	ch = diff.Diff(len(ctt.From), len(ctt.To), &ctt)
	if len(ch) != 0 {
		return false
	}
	return true
}

type TableTransition struct {
	From, To []*Table
}

func (d TableTransition) Equal(i, j int) bool {
	return d.From[i].EqualTo(d.To[j])
}

func (d TableTransition) Diff() []diff.Change {
	return diff.Diff(len(d.From), len(d.To), d)
}

func (c1 *Column) EqualTo(c2 *Column) bool {
	if c1 == c2 {
		return true
	}
	if c1.Name != c2.Name ||
		c1.Type != c2.Type ||
		c1.Nullable != c2.Nullable ||
		c1.PrimaryKey != c2.PrimaryKey ||
		c1.Default.String != c2.Default.String ||
		c1.Default.Valid != c2.Default.Valid {
		return false
	}
	return true
}

type ColumnTransition struct {
	From, To []*Column
}

func (d ColumnTransition) Equal(i, j int) bool {
	return d.From[i].EqualTo(d.To[j])
}

func (d ColumnTransition) Diff() []diff.Change {
	return diff.Diff(len(d.From), len(d.To), d)
}

func (idx1 *Index) EqualTo(idx2 *Index) bool {
	if idx1 == idx2 {
		return true
	}
	if idx1.Name != idx2.Name ||
		idx1.IsPrimary != idx2.IsPrimary ||
		idx1.IsUnique != idx2.IsUnique ||
		idx1.IsClustered != idx2.IsClustered ||
		idx1.MethodName != idx2.MethodName ||
		idx1.Def != idx2.Def ||
		strings.Join(idx1.Columns, ",") != strings.Join(idx2.Columns, ",") ||
		idx1.Concurrently != idx2.Concurrently ||
		idx1.ColDef != idx2.ColDef ||
		idx1.With != idx2.With ||
		idx1.Tablespace != idx2.Tablespace ||
		idx1.Where != idx2.Where {
		return false
	}
	return true
}

type IndexTransition struct {
	From, To []*Index
}

func (d IndexTransition) Equal(i, j int) bool {
	return d.From[i].EqualTo(d.To[j])
}

func (d IndexTransition) Diff() []diff.Change {
	return diff.Diff(len(d.From), len(d.To), d)
}

func (c1 *Constraint) EqualTo(c2 *Constraint) bool {
	if c1 == c2 {
		return true
	}
	if c1.Name != c2.Name ||
		c1.Type != c2.Type ||
		c1.Def != c2.Def ||
		c1.Check != c2.Check ||
		c1.OnDelete != c2.OnDelete ||
		*c1.ReferenceTable != *c2.ReferenceTable ||
		strings.Join(c1.Columns, ",") != strings.Join(c2.Columns, ",") ||
		strings.Join(c1.ReferenceColumns, ",") != strings.Join(c2.ReferenceColumns, ",") {
		return false
	}
	return true
}

type ConstraintTransition struct {
	From, To []*Constraint
}

func (d ConstraintTransition) Equal(i, j int) bool {
	return d.From[i].EqualTo(d.To[j])
}

func (d ConstraintTransition) Diff() []diff.Change {
	return diff.Diff(len(d.From), len(d.To), d)
}

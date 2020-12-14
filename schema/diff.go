package schema

import "github.com/covrom/diff"

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

type ColumnTransition struct {
	From, To []*Column
}

func (d ColumnTransition) Equal(i, j int) bool {
	return d.From[i] == d.To[j]
}

type IndexTransition struct {
	From, To []*Index
}

func (d IndexTransition) Equal(i, j int) bool {
	return d.From[i] == d.To[j]
}

type ConstraintTransition struct {
	From, To []*Constraint
}

func (d ConstraintTransition) Equal(i, j int) bool {
	return d.From[i] == d.To[j]
}

package schema

import "fmt"

var PatchDropDisable bool = false

type PatchTable struct {
	from, to    *Table
	columns     []*PatchColumn
	indexes     []*PatchIndex
	constraints []*PatchConstraint
}

func (t *PatchTable) GenerateSQL() []string {
	if t.from != nil && t.to != nil {
		return t.alter()
	}
	if t.from == nil {
		return t.create()
	}
	return t.drop()
}
func (t *PatchTable) create() []string { return nil }
func (t *PatchTable) alter() []string  { return nil }
func (t *PatchTable) drop() []string {
	if PatchDropDisable {
		return nil
	}
	return []string{
		fmt.Sprintf("DROP TABLE IF EXISTS %s", t.from.Name),
	}
}

type PatchColumn struct {
	from, to  *Column
	tableName string
}

func (c *PatchColumn) GenerateSQL() []string {
	if c.from != nil && c.to != nil {
		return c.alter()
	}
	if c.from == nil {
		return c.create()
	}
	return c.drop()
}
func (c *PatchColumn) create() []string { return nil }
func (c *PatchColumn) alter() []string  { return nil }
func (c *PatchColumn) drop() []string {
	if PatchDropDisable {
		return nil
	}
	return []string{
		fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s", c.tableName, c.from.Name),
	}
}

type PatchIndex struct {
	from, to *Index
}

func (i *PatchIndex) GenerateSQL() []string {
	if i.from != nil && i.to != nil {
		return i.alter()
	}
	if i.from == nil {
		return i.create()
	}
	return i.drop()
}
func (i *PatchIndex) create() []string { return nil }
func (i *PatchIndex) alter() []string  { return nil }
func (i *PatchIndex) drop() []string {
	// always drop unused indexes
	return []string{
		fmt.Sprintf("DROP INDEX IF EXISTS %s", i.from.Name),
	}
}

type PatchConstraint struct {
	from, to  *Constraint
	tableName string
}

func (c *PatchConstraint) GenerateSQL() []string {
	if c.from != nil && c.to != nil {
		return c.alter()
	}
	if c.from == nil {
		return c.create()
	}
	return c.drop()
}
func (c *PatchConstraint) create() []string { return nil }
func (c *PatchConstraint) alter() []string  { return nil }
func (c *PatchConstraint) drop() []string {
	// always drop unused constraints
	return []string{
		fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", c.tableName, c.from.Name),
	}
}

type PatchRelation struct {
	from, to *Relation
}

func (r *PatchRelation) GenerateSQL() []string {
	if r.from != nil && r.to != nil {
		return r.alter()
	}
	if r.from == nil {
		return r.create()
	}
	return r.drop()
}
func (r *PatchRelation) create() []string { return nil }
func (r *PatchRelation) alter() []string  { return nil }
func (r *PatchRelation) drop() []string {
	// TODO:
	// declare r record;
	// begin
	//   for r in (
	// 	select constraint_name
	// 	from information_schema.table_constraints
	// 	where table_name='relationships'
	// 	and constraint_name like 'fk_%'
	//   ) loop
	//   execute CONCAT('ALTER TABLE "relationships" DROP CONSTRAINT '||r.constraint_name);
	//   end loop;
	// end;
	return nil
}

type PatchSchema struct {
	CurrentSchema string
	tables        []*PatchTable
	relations     []*PatchRelation
}

func (t *PatchSchema) GenerateSQL() (ret []string) {
	// TODO: using CurrentSchema
	for _, st := range t.tables {
		ret = append(ret, st.GenerateSQL()...)
	}
	for _, rt := range t.relations {
		ret = append(ret, rt.GenerateSQL()...)
	}
	return
}

func (s *PatchSchema) Build(from, to *Schema) {
	s.CurrentSchema = to.CurrentSchema
	s.tables = make([]*PatchTable, 0, len(from.Tables)+len(to.Tables))
	s.relations = make([]*PatchRelation, 0, len(from.Relations)+len(to.Relations))

	// drop or alter tables
	for _, t := range from.Tables {
		pt := &PatchTable{
			from: t,
		}
		rt, err := to.FindTableByName(t.Name)
		if err == nil {
			pt.to = rt
		}
		s.tables = append(s.tables, pt)
		for _, c := range t.Columns {
			pc := &PatchColumn{
				tableName: t.Name,
				from:      c,
			}
			if rt != nil {
				rc, err := rt.FindColumnByName(c.Name)
				if err == nil {
					pc.to = rc
				}
			}
			pt.columns = append(pt.columns, pc)
		}
		for _, idx := range t.Indexes {
			pi := &PatchIndex{
				from: idx,
			}
			if rt != nil {
				ri, err := rt.FindIndexByName(idx.Name)
				if err == nil {
					pi.to = ri
				}
			}
			pt.indexes = append(pt.indexes, pi)
		}
		for _, c := range t.Constraints {
			pc := &PatchConstraint{
				tableName: t.Name,
				from:      c,
			}
			if rt != nil {
				rc, err := rt.FindConstraintByName(c.Name)
				if err == nil {
					pc.to = rc
				}
			}
			pt.constraints = append(pt.constraints, pc)
		}
	}
	// create tables
	for _, rt := range to.Tables {
		fnd := false
		for _, t := range s.tables {
			if t.to == nil {
				continue
			}
			if t.to.Name == rt.Name {
				fnd = true
				break
			}
		}
		if fnd {
			continue
		}
		pt := &PatchTable{to: rt}
		s.tables = append(s.tables, pt)
		for _, c := range rt.Columns {
			pc := &PatchColumn{
				tableName: rt.Name,
				to:        c,
			}
			pt.columns = append(pt.columns, pc)
		}
		for _, idx := range rt.Indexes {
			pi := &PatchIndex{
				to: idx,
			}
			pt.indexes = append(pt.indexes, pi)
		}
		for _, c := range rt.Constraints {
			pc := &PatchConstraint{
				tableName: rt.Name,
				to:        c,
			}
			pt.constraints = append(pt.constraints, pc)
		}
	}

	// drop or alter relations
	for _, r := range from.Relations {
		pt := &PatchRelation{
			from: r,
		}
		rt, err := to.FindRelation(r.Columns, r.ParentColumns)
		if err == nil {
			pt.to = rt
		}
		s.relations = append(s.relations, pt)
	}
	// create relations
	for _, r := range to.Relations {
		pt := &PatchRelation{
			to: r,
		}
		_, err := from.FindRelation(r.Columns, r.ParentColumns)
		if err != nil {
			s.relations = append(s.relations, pt)
		}
	}
}

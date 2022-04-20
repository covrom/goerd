package schema

import (
	"fmt"
	"strings"
)

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

func (t *PatchTable) create() []string {
	if t.to.Type == "" {
		t.to.Type = "TABLE"
	}
	if t.to.Type != "TABLE" {
		return []string{
			fmt.Sprintf("CREATE %s %s AS (\n%s\n)", t.to.Type, t.to.Name, strings.TrimRight(t.to.Def, ";")),
		}
	}

	sb := &strings.Builder{}
	fmt.Fprint(sb, "CREATE TABLE ", t.to.Name, " (\n")
	crlf := false
	for _, c := range t.columns {
		if crlf {
			sb.WriteString(",\n")
		} else {
			crlf = true
		}
		sb.WriteString(c.create()[0])
	}
	for _, cs := range t.constraints {
		if crlf {
			sb.WriteString(",\n")
		} else {
			crlf = true
		}
		sb.WriteString(cs.create()[0])
	}
	fmt.Fprint(sb, ")")

	ret := []string{sb.String()}

	for _, idx := range t.indexes {
		idx.to.Table = &t.to.Name
		ret = append(ret, idx.create()...)
	}

	return ret
}

func (t *PatchTable) alter() []string {
	ret := []string{}
	for _, c := range t.columns {
		if c.from == nil {
			ret = append(ret, c.create()...)
		} else if c.to == nil {
			ret = append(ret, c.drop()...)
		} else {
			ret = append(ret, c.alter()...)
		}
	}
	for _, idx := range t.indexes {
		if idx.from == nil {
			ret = append(ret, idx.create()...)
		} else if idx.to == nil {
			ret = append(ret, idx.drop()...)
		} else {
			ret = append(ret, idx.alter()...)
		}
	}
	for _, ctr := range t.constraints {
		if ctr.from == nil {
			ret = append(ret, ctr.create()...)
		} else if ctr.to == nil {
			ret = append(ret, ctr.drop()...)
		} else {
			ret = append(ret, ctr.alter()...)
		}
	}
	return ret
}

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
	newTable  bool
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

func (c *PatchColumn) create() []string {
	sb := &strings.Builder{}
	if !c.newTable {
		fmt.Fprintf(sb, "ALTER TABLE %s ADD COLUMN ", c.tableName)
	}
	fmt.Fprint(sb, c.to.Name, " ", c.to.Type)
	if !c.to.Nullable {
		fmt.Fprint(sb, " NOT NULL")
	}
	if c.to.Default.Valid {
		fmt.Fprint(sb, " DEFAULT ", c.to.Default.String)
	}
	if c.to.PrimaryKey {
		fmt.Fprint(sb, " PRIMARY KEY")
	}
	return []string{sb.String()}
}

func (c *PatchColumn) alter() []string {
	ret := []string{}
	if c.from.Default.String != c.to.Default.String && (c.from.Default.Valid || c.to.Default.Valid) {
		if c.to.Default.Valid {
			ret = append(ret, fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
				c.tableName, c.to.Name, c.to.Default.String,
			))
		} else {
			ret = append(ret, fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
				c.tableName, c.to.Name,
			))
		}
	}
	if c.from.Nullable != c.to.Nullable {
		if c.to.Nullable {
			ret = append(ret, fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
				c.tableName, c.to.Name,
			))
		} else {
			ret = append(ret, fmt.Sprintf(
				"ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
				c.tableName, c.to.Name,
			))
		}
	}
	if c.from.Type != c.to.Type {
		ret = append(ret, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			c.tableName, c.to.Name, c.to.Type,
		))
	}
	return ret
}

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

func createIndexDDL(idx *Index) string {
	sb := &strings.Builder{}
	fmt.Fprint(sb, "CREATE")
	if idx.IsUnique {
		fmt.Fprint(sb, " UNIQUE")
	}
	fmt.Fprint(sb, " INDEX")
	if idx.Concurrently {
		fmt.Fprint(sb, " CONCURRENTLY")
	}
	fmt.Fprintf(sb, " %s ON %s",
		idx.Name, *idx.Table)
	if len(idx.MethodName) > 0 {
		fmt.Fprint(sb, " USING ", idx.MethodName)
	}
	sb.WriteByte('(')
	if len(idx.ColDef) > 0 {
		sb.WriteString(idx.ColDef)
	} else {
		fmt.Fprint(sb, strings.Join(idx.Columns, ", "))
	}
	sb.WriteByte(')')
	if len(idx.With) > 0 {
		fmt.Fprint(sb, " WITH ", idx.With)
	}
	if len(idx.Tablespace) > 0 {
		fmt.Fprint(sb, " TABLESPACE ", idx.Tablespace)
	}
	if len(idx.Where) > 0 {
		fmt.Fprint(sb, " WHERE ", idx.Where)
	}
	return sb.String()
}

func (i *PatchIndex) create() []string {
	return []string{createIndexDDL(i.to)}
}
func (i *PatchIndex) alter() []string {
	if i.from.MethodName == "" {
		i.from.MethodName = "btree"
	}
	if i.to.MethodName == "" {
		i.to.MethodName = "btree"
	}
	if strings.EqualFold(createIndexDDL(i.from), createIndexDDL(i.to)) {
		return nil
	}
	return append(i.drop(), i.create()...)
}
func (i *PatchIndex) drop() []string {
	// always drop unused indexes
	return []string{
		fmt.Sprintf("DROP INDEX IF EXISTS %s", i.from.Name),
	}
}

type PatchConstraint struct {
	from, to  *Constraint
	tableName string
	newTable  bool
}

func (c *PatchConstraint) GenerateSQL() []string {
	if c.from != nil && c.to != nil {
		c.from.Table = &c.tableName
		c.to.Table = &c.tableName
		return c.alter()
	}
	if c.from == nil {
		c.to.Table = &c.tableName
		return c.create()
	}
	c.from.Table = &c.tableName
	return c.drop()
}

func createConstraintDDL(ctr *Constraint, newTable bool) string {
	sb := &strings.Builder{}
	if !newTable {
		fmt.Fprintf(sb, "ALTER TABLE %s ADD ", *ctr.Table)
	}
	fmt.Fprint(sb, "CONSTRAINT ", ctr.Name)
	if len(ctr.Check) > 0 {
		fmt.Fprint(sb, " CHECK (", ctr.Check, ")")
	}
	switch ctr.Type {
	case TypeFK:
		if ctr.ReferenceTable == nil {
			fmt.Fprint(sb, " FOREIGN KEY (", strings.Join(ctr.Columns, ", "), ")")
			fmt.Fprintf(sb, " REFERENCES %s (%s)", *ctr.ReferenceTable, strings.Join(ctr.ReferenceColumns, ", "))
			if len(ctr.OnDelete) > 0 {
				fmt.Fprint(sb, " ON DELETE ", ctr.OnDelete)
			}
		}
	case TypePK:
		fmt.Fprint(sb, " PRIMARY KEY (", strings.Join(ctr.Columns, ", "), ")")
	case TypeUQ:
		fmt.Fprint(sb, " UNIQUE (", strings.Join(ctr.Columns, ", "), ")")
	}
	return sb.String()
}

func (c *PatchConstraint) create() []string {
	return []string{createConstraintDDL(c.to, c.newTable)}
}

func (c *PatchConstraint) alter() []string {
	if strings.EqualFold(createConstraintDDL(c.from, c.newTable),
		createConstraintDDL(c.to, c.newTable)) {
		return nil
	}
	return append(c.drop(), c.create()...)
}

func (c *PatchConstraint) drop() []string {
	if c.to == nil && c.from.Type == TypePK {
		// pk not drop
		return nil
	}
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

func createRelationDDL(r *Relation) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (", r.Table.Name, r.Name)
	for i, c := range r.Columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(c.Name)
	}
	fmt.Fprintf(sb, ") REFERENCES %s (", r.ParentTable.Name)
	for i, c := range r.ParentColumns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(c.Name)
	}
	sb.WriteByte(')')
	return sb.String()
}

func (r *PatchRelation) create() []string {
	return []string{createRelationDDL(r.to)}
}
func (r *PatchRelation) alter() []string {
	if strings.EqualFold(createRelationDDL(r.from),
		createRelationDDL(r.to)) {
		return nil
	}
	return append(r.drop(), r.create()...)
}
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
		if rt != nil {
			for _, c := range rt.Columns {
				pc := &PatchColumn{
					tableName: t.Name,
					to:        c,
				}
				tc, err := t.FindColumnByName(c.Name)
				fnd := false
				if err == nil {
					pc.from = tc

					for _, v := range pt.columns {
						if v.from == pc.from &&
							v.to == pc.to {
							fnd = true
							break
						}
					}
				}
				if !fnd {
					pt.columns = append(pt.columns, pc)
				}
			}
		}
		for _, idx := range t.Indexes {
			if idx.Table == nil {
				idx.Table = &t.Name
			}
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

		if rt != nil {
			for _, idx := range rt.Indexes {
				if idx.Table == nil {
					idx.Table = &rt.Name
				}
				pi := &PatchIndex{
					to: idx,
				}
				ti, err := t.FindIndexByName(idx.Name)
				fnd := false
				if err == nil {
					pi.from = ti

					for _, v := range pt.indexes {
						if v.from == pi.from &&
							v.to == pi.to {
							fnd = true
							break
						}
					}
				}
				if !fnd {
					pt.indexes = append(pt.indexes, pi)
				}
			}
		}

		for _, c := range t.Constraints {
			if c.Table == nil {
				c.Table = &t.Name
			}
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

		if rt != nil {
			for _, c := range rt.Constraints {
				if c.Table == nil {
					c.Table = &rt.Name
				}
				pc := &PatchConstraint{
					tableName: rt.Name,
					to:        c,
				}
				tc, err := t.FindConstraintByName(c.Name)
				fnd := false
				if err == nil {
					pc.from = tc

					for _, v := range pt.constraints {
						if v.from == pc.from &&
							v.to == pc.to {
							fnd = true
							break
						}
					}
				}
				if !fnd {
					pt.constraints = append(pt.constraints, pc)
				}
			}
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
				newTable:  true,
			}
			pt.columns = append(pt.columns, pc)
		}
		for _, idx := range rt.Indexes {
			if idx.Table == nil {
				idx.Table = &rt.Name
			}
			pi := &PatchIndex{
				to: idx,
			}
			pt.indexes = append(pt.indexes, pi)
		}
		for _, c := range rt.Constraints {
			if c.Table == nil {
				c.Table = &rt.Name
			}
			pc := &PatchConstraint{
				tableName: rt.Name,
				to:        c,
				newTable:  true,
			}
			pt.constraints = append(pt.constraints, pc)
		}
	}

	// drop or alter relations
	for _, r := range from.Relations {
		pt := &PatchRelation{
			from: r,
		}
		rt, err := to.FindRelation(r.Table.Name, r.Columns, r.ParentColumns)
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
		_, err := from.FindRelation(r.Table.Name, r.Columns, r.ParentColumns)
		if err != nil {
			s.relations = append(s.relations, pt)
		}
	}
}

package postgres

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/covrom/goerd/drivers/postgres/idxscan"
	"github.com/covrom/goerd/schema"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var reFK = regexp.MustCompile(`(?is)FOREIGN\s+KEY\s*\((.+)\)\s*REFERENCES\s+(\S+)\s*\((.+)\)(\s*ON\s+DELETE\s+((CASCADE)|(RESTRICT)|(NO\s+ACTION)|(SET\s+NULL)|(SET\s+DEFAULT)))?`)
var reChk = regexp.MustCompile(`(?is)CHECK\s+\((.+)\)\s*$`)

// Postgres struct
type Postgres struct {
	db *sql.DB
}

// New return new Postgres
func New(db *sql.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

// Analyze PostgreSQL database schema
func (p *Postgres) Analyze(s *schema.Schema) error {
	// current schema
	var currentSchema string
	schemaRows, err := p.db.Query(`SELECT current_schema()`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer schemaRows.Close()
	for schemaRows.Next() {
		err := schemaRows.Scan(&currentSchema)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	s.CurrentSchema = currentSchema

	// search_path
	var searchPaths string
	pathRows, err := p.db.Query(`SHOW search_path`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer pathRows.Close()
	for pathRows.Next() {
		err := pathRows.Scan(&searchPaths)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	s.SearchPaths = strings.Split(searchPaths, ", ")

	fullTableNames := []string{}

	// tables
	tableRows, err := p.db.Query(`
SELECT
    cls.oid AS oid,
    cls.relname AS table_name,
    CASE
        WHEN cls.relkind IN ('r', 'p') THEN 'TABLE'
        WHEN cls.relkind = 'v' THEN 'VIEW'
        WHEN cls.relkind = 'm' THEN 'MATERIALIZED VIEW'
        WHEN cls.relkind = 'f' THEN 'FOREIGN TABLE'
    END AS table_type,
    ns.nspname AS table_schema,
    descr.description AS table_comment
FROM pg_class AS cls
INNER JOIN pg_namespace AS ns ON cls.relnamespace = ns.oid
LEFT JOIN pg_description AS descr ON cls.oid = descr.objoid AND descr.objsubid = 0
WHERE ns.nspname NOT IN ('pg_catalog', 'information_schema')
AND cls.relkind IN ('r', 'p', 'v', 'f', 'm')
ORDER BY oid`)
	if err != nil {
		return errors.WithStack(err)
	}
	defer tableRows.Close()

	relations := []*schema.Relation{}

	tables := []*schema.Table{}
	for tableRows.Next() {
		var (
			tableOid     uint64
			tableName    string
			tableType    string
			tableSchema  string
			tableComment sql.NullString
		)
		err := tableRows.Scan(&tableOid, &tableName, &tableType, &tableSchema, &tableComment)
		if err != nil {
			return errors.WithStack(err)
		}

		name := tableName
		if tableSchema != currentSchema {
			name = fmt.Sprintf("%s.%s", tableSchema, tableName)
		}

		fullTableNames = append(fullTableNames, fmt.Sprintf("%s.%s", tableSchema, tableName))

		table := &schema.Table{
			Name:    name,
			Type:    tableType,
			Comment: tableComment.String,
		}

		// (materialized) view definition
		if tableType == "VIEW" || tableType == "MATERIALIZED VIEW" {
			viewDefRows, err := p.db.Query(`SELECT pg_get_viewdef($1::oid);`, tableOid)
			if err != nil {
				return errors.WithStack(err)
			}
			defer viewDefRows.Close()
			for viewDefRows.Next() {
				var tableDef sql.NullString
				err := viewDefRows.Scan(&tableDef)
				if err != nil {
					return errors.WithStack(err)
				}
				//  fmt.Sprintf("CREATE %s %s AS (\n%s\n)", tableType, tableName, strings.TrimRight(tableDef.String, ";"))
				table.Def = strings.TrimRight(tableDef.String, ";")
			}
		}

		// constraints
		constraintRows, err := p.db.Query(p.queryForConstraints(), tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer constraintRows.Close()

		constraints := []*schema.Constraint{}

		for constraintRows.Next() {
			var (
				constraintName                 string
				constraintDef                  string
				constraintType                 string
				constraintReferenceTable       sql.NullString
				constraintColumnNames          []sql.NullString
				constraintReferenceColumnNames []sql.NullString
				constraintComment              sql.NullString
			)
			err = constraintRows.Scan(&constraintName, &constraintDef, &constraintType, &constraintReferenceTable, pq.Array(&constraintColumnNames), pq.Array(&constraintReferenceColumnNames), &constraintComment)
			if err != nil {
				return errors.WithStack(err)
			}
			prt := (*string)(nil)
			if constraintReferenceTable.Valid && constraintReferenceTable.String != "" {
				prt = &constraintReferenceTable.String
			}

			constraint := &schema.Constraint{
				Name:             constraintName,
				Type:             convertConstraintType(constraintType),
				Def:              constraintDef,
				Table:            &table.Name,
				Columns:          arrayRemoveNull(constraintColumnNames),
				ReferenceTable:   prt,
				ReferenceColumns: arrayRemoveNull(constraintReferenceColumnNames),
				Comment:          constraintComment.String,
			}

			ss := reChk.FindStringSubmatch(constraintDef)
			if len(ss) > 0 {
				constraint.Check = strings.TrimSpace(ss[1])
			}

			if constraintType == "f" {
				ss = reFK.FindStringSubmatch(constraintDef)
				if len(ss) > 0 {
					constraint.OnDelete = strings.TrimSpace(ss[5])
				}
				relation := &schema.Relation{
					Table:    table,
					OnDelete: constraint.OnDelete,
					Def:      constraintDef,
				}
				relations = append(relations, relation)
			} else {
				constraints = append(constraints, constraint)
			}
		}
		table.Constraints = constraints

		// columns
		columnRows, err := p.db.Query(`
SELECT
    attr.attname AS column_name,
    pg_get_expr(def.adbin, def.adrelid) AS column_default,
    NOT (attr.attnotnull OR tp.typtype = 'd' AND tp.typnotnull) AS is_nullable,
    CASE
        WHEN 'character varying'::regtype = ANY(ARRAY[attr.atttypid, tp.typelem]) THEN
            REPLACE(format_type(attr.atttypid, attr.atttypmod), 'character varying', 'varchar')
		WHEN 'timestamp with time zone'::regtype = ANY(ARRAY[attr.atttypid, tp.typelem]) THEN
            REPLACE(format_type(attr.atttypid, attr.atttypmod), 'timestamp with time zone', 'timestamptz')
        ELSE format_type(attr.atttypid, attr.atttypmod)
    END AS data_type,
    descr.description AS comment
FROM pg_attribute AS attr
INNER JOIN pg_type AS tp ON attr.atttypid = tp.oid
LEFT JOIN pg_attrdef AS def ON attr.attrelid = def.adrelid AND attr.attnum = def.adnum
LEFT JOIN pg_description AS descr ON attr.attrelid = descr.objoid AND attr.attnum = descr.objsubid
WHERE
    attr.attnum > 0
AND NOT attr.attisdropped
AND attr.attrelid = $1::oid
ORDER BY attr.attnum;
`, tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer columnRows.Close()

		columns := []*schema.Column{}
		for columnRows.Next() {
			var (
				columnName    string
				columnDefault sql.NullString
				isNullable    bool
				dataType      string
				columnComment sql.NullString
			)
			err = columnRows.Scan(&columnName, &columnDefault, &isNullable, &dataType, &columnComment)
			if err != nil {
				return errors.WithStack(err)
			}
			column := &schema.Column{
				Name:     columnName,
				Type:     dataType,
				Nullable: isNullable,
				Default:  columnDefault,
				Comment:  columnComment.String,
			}
			// find in pk's
			for _, cstr := range constraints {
				if cstr.Type != schema.TypePK {
					continue
				}
				for _, cscol := range cstr.Columns {
					if cscol == columnName {
						column.PrimaryKey = true
						break
					}
				}
			}
			columns = append(columns, column)
		}
		table.Columns = columns

		// indexes
		indexRows, err := p.db.Query(p.queryForIndexes(), tableOid)
		if err != nil {
			return errors.WithStack(err)
		}
		defer indexRows.Close()

		indexes := []*schema.Index{}
		for indexRows.Next() {
			var (
				indexName        string
				indexDef         string
				indisprimary     bool
				indisunique      bool
				indisclustered   bool
				amname           string
				indexColumnNames []sql.NullString
				indexComment     sql.NullString
			)
			err = indexRows.Scan(&indexName, &indisprimary,
				&indisunique, &indisclustered,
				&amname, &indexDef,
				pq.Array(&indexColumnNames), &indexComment)
			if err != nil {
				return errors.WithStack(err)
			}
			idxprs := idxscan.ParseCreateIndex(indexDef)
			index := &schema.Index{
				Name:         indexName,
				IsClustered:  indisclustered,
				IsPrimary:    indisprimary,
				IsUnique:     indisunique,
				MethodName:   amname,
				Def:          indexDef,
				Table:        &table.Name,
				Concurrently: idxprs.Concurrently,
				ColDef:       idxprs.ColDef,
				With:         idxprs.With,
				Tablespace:   idxprs.Tablespace,
				Where:        idxprs.Where,
				Columns:      arrayRemoveNull(indexColumnNames),
				Comment:      indexComment.String,
			}
			if strings.ReplaceAll(idxprs.ColDef, " ", "") == "("+strings.Join(index.Columns, ",")+")" {
				index.ColDef = ""
			}
			if index.MethodName == "btree" &&
				index.Where == "" && index.With == "" &&
				!index.Concurrently && !index.IsClustered {
				csfnd := false
				for _, cs := range table.Constraints {
					if cs.Name == index.Name &&
						(cs.Type == schema.TypeUQ || cs.Type == schema.TypePK) {
						csfnd = true
						break
					}
				}
				if csfnd {
					continue
				}
			}
			indexes = append(indexes, index)
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	s.Tables = tables

	// Relations
	for _, r := range relations {
		result := reFK.FindStringSubmatch(r.Def)
		if len(result) == 0 {
			continue
		}
		strColumns := []string{}
		for _, c := range strings.Split(result[1], ", ") {
			strColumns = append(strColumns, strings.ReplaceAll(c, `"`, ""))
		}
		strParentTable := strings.ReplaceAll(result[2], `"`, "")
		strParentColumns := []string{}
		for _, c := range strings.Split(result[3], ", ") {
			strParentColumns = append(strParentColumns, strings.ReplaceAll(c, `"`, ""))
		}
		for _, c := range strColumns {
			column, err := r.Table.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.Columns = append(r.Columns, column)
			column.ParentRelations = append(column.ParentRelations, r)
		}

		dn, err := detectFullTableName(strParentTable, s.SearchPaths, fullTableNames)
		if err != nil {
			return err
		}
		strParentTable = dn
		parentTable, err := s.FindTableByName(strParentTable)
		if err != nil {
			return err
		}
		r.ParentTable = parentTable
		for _, c := range strParentColumns {
			column, err := parentTable.FindColumnByName(c)
			if err != nil {
				return err
			}
			r.ParentColumns = append(r.ParentColumns, column)
			column.ChildRelations = append(column.ChildRelations, r)
		}
	}

	s.Relations = relations

	return nil
}

func (p *Postgres) queryForConstraints() string {
	return `
SELECT
  cons.conname AS name,
  CASE WHEN cons.contype = 't' THEN pg_get_triggerdef(trig.oid)
        ELSE pg_get_constraintdef(cons.oid)
  END AS def,
  cons.contype AS type,
  fcls.relname,
  ARRAY_AGG(attr.attname),
  ARRAY_AGG(fattr.attname),
  descr.description AS comment
FROM pg_constraint AS cons
LEFT JOIN pg_trigger AS trig ON trig.tgconstraint = cons.oid AND NOT trig.tgisinternal
LEFT JOIN pg_class AS fcls ON cons.confrelid = fcls.oid
LEFT JOIN pg_attribute AS attr ON attr.attrelid = cons.conrelid
LEFT JOIN pg_attribute AS fattr ON fattr.attrelid = cons.confrelid
LEFT JOIN pg_description AS descr ON cons.oid = descr.objoid
WHERE
	cons.conrelid = $1::oid
AND (cons.conkey IS NULL OR attr.attnum = ANY(cons.conkey))
AND (cons.confkey IS NULL OR fattr.attnum = ANY(cons.confkey))
GROUP BY cons.conindid, cons.conname, cons.contype, cons.oid, trig.oid, fcls.relname, descr.description
ORDER BY cons.conindid, cons.conname`
}

// arrayRemoveNull
func arrayRemoveNull(in []sql.NullString) []string {
	out := []string{}
	for _, i := range in {
		if i.Valid {
			out = append(out, i.String)
		}
	}
	return out
}

func (p *Postgres) queryForIndexes() string {
	return `
SELECT
  cls.relname AS indexname,
  idx.indisprimary,
  idx.indisunique,
  idx.indisclustered,
  am.amname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  ARRAY_AGG(attr.attname),
  descr.description AS comment
FROM pg_index AS idx
INNER JOIN pg_class AS cls ON idx.indexrelid = cls.oid
INNER JOIN pg_attribute AS attr ON idx.indexrelid = attr.attrelid
LEFT JOIN pg_description AS descr ON idx.indexrelid = descr.objoid
LEFT JOIN pg_am am ON am.oid=cls.relam
WHERE idx.indrelid = $1::oid
GROUP BY cls.relname, idx.indexrelid, descr.description, idx.indisprimary, idx.indisunique, idx.indisclustered, am.amname
ORDER BY idx.indexrelid`
}

func detectFullTableName(name string, searchPaths, fullTableNames []string) (string, error) {
	if strings.Contains(name, ".") {
		return name, nil
	}
	fns := []string{}
	for _, n := range fullTableNames {
		if strings.HasSuffix(n, name) {
			for _, p := range searchPaths {
				// TODO: Support $user
				if n == fmt.Sprintf("%s.%s", p, name) {
					fns = append(fns, n)
				}
			}
		}
	}
	if len(fns) != 1 {
		return "", errors.Errorf("can not detect table name: %s", name)
	}
	return fns[0], nil
}

func convertConstraintType(t string) string {
	switch t {
	case "p":
		return schema.TypePK
	case "u":
		return schema.TypeUQ
	case "f":
		return schema.TypeFK
	case "c":
		return "CHECK"
	case "t":
		return "TRIGGER"
	default:
		return t
	}
}

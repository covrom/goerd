package postgres

const (
	qTables = `
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
ORDER BY oid`

	qContstraints = `
SELECT
  cons.conname AS name,
  CASE WHEN cons.contype = 't' THEN pg_get_triggerdef(trig.oid)
        ELSE pg_get_constraintdef(cons.oid)
  END AS def,
  cons.contype AS type,
  fcls.relname,
  array_to_json(ARRAY_AGG(attr.attname)) as attnm,
  array_to_json(ARRAY_AGG(fattr.attname)) as fattnm,
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

	qColumns = `
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
`

	qIndexes = `
SELECT
  cls.relname AS indexname,
  idx.indisprimary,
  idx.indisunique,
  idx.indisclustered,
  am.amname,
  pg_get_indexdef(idx.indexrelid) AS indexdef,
  array_to_json(ARRAY_AGG(attr.attname)) as attnm,
  descr.description AS comment
FROM pg_index AS idx
INNER JOIN pg_class AS cls ON idx.indexrelid = cls.oid
INNER JOIN pg_attribute AS attr ON idx.indexrelid = attr.attrelid
LEFT JOIN pg_description AS descr ON idx.indexrelid = descr.objoid
LEFT JOIN pg_am am ON am.oid=cls.relam
WHERE idx.indrelid = $1::oid
GROUP BY cls.relname, idx.indexrelid, descr.description, idx.indisprimary, idx.indisunique, idx.indisclustered, am.amname
ORDER BY idx.indexrelid`
)

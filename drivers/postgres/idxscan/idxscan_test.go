package idxscan

import (
	"reflect"
	"testing"
)

func TestParseCreateIndex(t *testing.T) {
	s := `  CREATE	UNiQUE
	INDEX	uk_candidate_recruiter_info_email	ON public.candidate
	USING btree
	(		sdih COLLATE "ru_RU" int4 ASC NULLS LAST,
	asfg COLLATE "ru_RU" DESC,
	ertny,
	((recruiter_info ->>
	'email'::text))
	)
 
	WITH a= b, c = d,e=f
	TABLESPACE f
	WHERE as;roug (hblkas) bfdvvkjhb (3li54f9q3er opfbg78 9)erg`

	ss := ParseCreateIndex(s)
	eq := IndexDef{
		Name:         "uk_candidate_recruiter_info_email",
		Table:        "public.candidate",
		Unique:       true,
		Concurrently: false,
		UsingType:    "btree",
		ColDef:       "(\t\tsdih COLLATE \"ru_RU\" int4 ASC NULLS LAST,\n asfg COLLATE \"ru_RU\" DESC,\n ertny,\n ((recruiter_info ->>\n 'email'::text))\n )",
		With:         "a= b, c = d,e=f",
		Tablespace:   "f",
		Where:        "as;roug (hblkas) bfdvvkjhb (3li54f9q3er opfbg78 9)erg",
	}

	if !reflect.DeepEqual(ss, eq) {
		t.Errorf("%#v", ss)
	}
}

func TestParseCreateIndex_NoWithAndTimestampWithTimezone(t *testing.T) {
	s := `CREATE UNIQUE INDEX \"IX_customer_external_id\"
	ON public.customer
	USING btree (external_id)
	WHERE (created > '2020-09-27 20:00:00.000000'::timestamp with time zone)`

	ss := ParseCreateIndex(s)
	eq := IndexDef{
		Name:         "\\\"IX_customer_external_id\\\"",
		Table:        "public.customer",
		Unique:       true,
		Concurrently: false,
		UsingType:    "btree",
		ColDef:       "(external_id)",
		With:         "",
		Tablespace:   "",
		Where:        "(created > '2020-09-27 20:00:00.000000'::timestamp with time zone)",
	}

	if !reflect.DeepEqual(ss, eq) {
		t.Errorf("%#v", ss)
	}
}

func TestParseCreateIndex_WithAndTimestampWithoutTimezone(t *testing.T) {
	s := `CREATE UNIQUE INDEX \"IX_customer_external_id\"
	ON public.customer
	USING btree (external_id)
	WITH a= b
	WHERE (created > '2020-09-27 20:00:00.000000'::timestamp without time zone)`

	ss := ParseCreateIndex(s)
	eq := IndexDef{
		Name:         "\\\"IX_customer_external_id\\\"",
		Table:        "public.customer",
		Unique:       true,
		Concurrently: false,
		UsingType:    "btree",
		ColDef:       "(external_id)",
		With:         "a= b",
		Tablespace:   "",
		Where:        "(created > '2020-09-27 20:00:00.000000'::timestamp without time zone)",
	}

	if !reflect.DeepEqual(ss, eq) {
		t.Errorf("%#v", ss)
	}
}

func TestParseCreateIndex_NoWhere(t *testing.T) {
	s := `CREATE UNIQUE INDEX \"IX_customer_external_id\"
	ON public.customer
	USING btree (external_id)
	WITH a= b`

	ss := ParseCreateIndex(s)
	eq := IndexDef{
		Name:         "\\\"IX_customer_external_id\\\"",
		Table:        "public.customer",
		Unique:       true,
		Concurrently: false,
		UsingType:    "btree",
		ColDef:       "(external_id)",
		With:         "a= b",
		Tablespace:   "",
		Where:        "",
	}

	if !reflect.DeepEqual(ss, eq) {
		t.Errorf("%#v", ss)
	}
}

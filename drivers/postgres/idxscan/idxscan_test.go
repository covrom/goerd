package idxscan

import (
	"reflect"
	"testing"
)

const testingQ = `  CREATE	UNiQUE
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

func TestParseCreateIndex(t *testing.T) {
	ss := ParseCreateIndex(testingQ)
	eq := IndexDef{
		Name:         "uk_candidate_recruiter_info_email",
		Table:        "public.candidate",
		Unique:       true,
		Concurrently: false,
		UsingType:    "btree",
		Tail:         "(\t\tsdih COLLATE \"ru_RU\" int4 ASC NULLS LAST,\n asfg COLLATE \"ru_RU\" DESC,\n ertny,\n ((recruiter_info ->>\n 'email'::text))\n )\n\n WITH a= b, c = d,e=f\n TABLESPACE f\n WHERE as;roug (hblkas) bfdvvkjhb (3li54f9q3er opfbg78 9)erg",
	}
	if !reflect.DeepEqual(ss, eq) {
		t.Errorf("%#v", ss)
	}
}

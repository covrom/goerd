package goerd

import (
	"fmt"
	"strings"
)

type Columner interface {
	Table() string
	Columns() []string
}

func ReplaceQuery(d Columner, idCols ...string) string {
	idcs := strings.Join(idCols, ",")
	cols := d.Columns()
	colsnoids := ColumnsWithout(cols, append(idCols, "created_at")...)
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s) ON CONFLICT(%s) DO UPDATE SET (%s)=(%s)`,
		d.Table(),
		strings.Join(cols, ","),
		Replacers(len(cols)),
		idcs,
		strings.Join(colsnoids, ","),
		strings.Join(Excluded(colsnoids), ","),
	)
}

func Replacers(cnt int) string {
	sb := &strings.Builder{}
	sb.Grow(cnt * 2)
	for i := 0; i < cnt; i++ {
		fmt.Fprintf(sb, "$%d", i+1)
		if i < cnt-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

func ColumnsWithout(cols []string, skip ...string) []string {
	if len(skip) == 0 {
		return cols
	}
	ret := make([]string, 0, len(cols))

	for _, c := range cols {
		fnd := false
		for _, v := range skip {
			if strings.EqualFold(c, v) {
				fnd = true
				break
			}
		}
		if !fnd {
			ret = append(ret, c)
		}
	}
	return ret
}

func Excluded(cols []string) []string {
	ret := make([]string, len(cols))
	for i, v := range cols {
		ret[i] = fmt.Sprintf("excluded.%s", v)
	}
	return ret
}

package idxscan

import (
	"regexp"
	"strings"
)

var crIdx = regexp.MustCompile(`(?is)\s*CREATE(\s+UNIQUE)?\s+INDEX(\s+CONCURRENTLY)?(\s+IF\s+NOT\s+EXISTS)?(\s+\S+)?\s+ON\s+(\S+)(\s+(USING)\s+(\S+))?(.+)`)

type IndexDef struct {
	Name         string
	Table        string
	Unique       bool
	Concurrently bool
	UsingType    string
	Tail         string
}

func ParseCreateIndex(s string) IndexDef {
	ss := crIdx.FindStringSubmatch(s)
	if ss == nil {
		return IndexDef{}
	}
	return IndexDef{
		Unique:       strings.EqualFold(strings.TrimSpace(ss[1]), "UNIQUE"),
		Concurrently: strings.EqualFold(strings.TrimSpace(ss[2]), "CONCURRENTLY"),
		Name:         strings.TrimSpace(ss[4]),
		Table:        strings.TrimSpace(ss[5]),
		UsingType:    strings.TrimSpace(ss[8]),
		Tail:         strings.TrimSpace(ss[9]),
	}
}

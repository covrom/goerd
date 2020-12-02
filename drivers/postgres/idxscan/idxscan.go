package idxscan

import (
	"regexp"
	"strings"
)

var crIdx = regexp.MustCompile(`(?is)\s*CREATE(\s+UNIQUE)?\s+INDEX(\s+CONCURRENTLY)?(\s+IF\s+NOT\s+EXISTS)?(\s+\S+)?\s+ON\s+(\S+)(\s+(USING)\s+(\S+))?(.+)`)
var fBrack = regexp.MustCompile(`(?is)(\((?:[^)(]+|\((?:[^)(]+|\((?:[^)(]+|\((?:[^)(]+|\((?:[^)(]+|\([^)(]*\))*\))*\))*\))*\))*\))(.*)`)

type IndexDef struct {
	Name         string
	Table        string
	Unique       bool
	Concurrently bool
	UsingType    string
	ColDef       string
	With         string
	Tablespace   string
	Where        string
}

func ParseCreateIndex(s string) IndexDef {
	ss := crIdx.FindStringSubmatch(s)
	if ss == nil {
		return IndexDef{}
	}
	idf := IndexDef{
		Unique:       strings.EqualFold(strings.TrimSpace(ss[1]), "UNIQUE"),
		Concurrently: strings.EqualFold(strings.TrimSpace(ss[2]), "CONCURRENTLY"),
		Name:         strings.TrimSpace(ss[4]),
		Table:        strings.TrimSpace(ss[5]),
		UsingType:    strings.TrimSpace(ss[8]),
	}
	tail := strings.TrimSpace(ss[9])
	brp := fBrack.FindStringSubmatch(tail)
	if brp != nil {
		idf.ColDef = strings.TrimSpace(brp[1])
		tail = brp[2]
		utail := strings.ToUpper(tail)
		idxwith := strings.Index(utail, "WITH")
		idxtbsp := strings.Index(utail, "TABLESPACE")
		idxwhere := strings.Index(utail, "WHERE")
		if idxwith >= 0 {
			idxto := len(tail)
			if idxwhere >= 0 {
				idxto = idxwhere
			}
			// WARN: idxtbsp<idxwhere
			if idxtbsp >= 0 {
				idxto = idxtbsp
			}
			idf.With = strings.TrimSpace(tail[idxwith+4 : idxto])
		}
		if idxtbsp >= 0 {
			idxto := len(tail)
			if idxwhere >= 0 {
				idxto = idxwhere
			}
			idf.Tablespace = strings.TrimSpace(tail[idxtbsp+10 : idxto])
		}
		if idxwhere >= 0 {
			idf.Where = strings.TrimSpace(tail[idxwhere+5:])
		}
	}
	return idf
}

package main

import (
	"strings"

	"github.com/go-leap/dev/go"
	"github.com/go-leap/fs"
)

type triplet struct{ a, b, c string }

type triplets []triplet

type trip struct {
	tname string
	vname string
	ctor  bool
	all   triplets
}

var (
	forStmts = trip{tname: "dr.Stmt", vname: "stmt", all: triplets{
		{"ex", "dr.StmtExecContext", "stmtExec"},
		{"qu", "dr.StmtQueryContext", "stmtQuery"},
	}}
	forConns = trip{tname: "dr.Conn", vname: "conn", ctor: false, all: triplets{
		{"bt", "dr.ConnBeginTx", "cbt"},
		{"pc", "dr.ConnPrepareContext", "cpc"},
		{"p", "dr.Pinger", "cp"},
		{"sr", "dr.SessionResetter", "csr"},
	}}
)

func main() {
	// forStmts.main()
	forConns.main()
}

func (me *trip) main() {
	rows := getIdxRows(len(me.all), len(me.all))
	var srclns []string
	for _, row := range rows {
		srcln := "case"
		for i, idx := range row {
			if i == 0 {
				srcln += " "
			} else {
				srcln += " && "
			}
			srcln += me.all[idx].a
		}
		srcln += ": "
		if me.ctor {
			srcln += "ctor = func(" + me.vname + " " + me.tname
			for i := range me.all {
				srcln += "," + me.all[i].c + " " + me.all[i].b
			}
			srcln += ")" + me.tname + "{"
		}
		srcln += "return &struct{" + me.tname
		for _, idx := range row {
			srcln += ";" + me.all[idx].b
		}

		srcln += "}{" + me.vname
		for _, idx := range row {
			srcln += "," + me.all[idx].c
		}

		srcln += "}"
		if me.ctor {
			srcln += "}"
		}
		srclns = append(srclns, srcln)
	}
	if e := ufs.WriteTextFile(udevgo.GopathSrc("github.com/go-leap/db/driver/proxy/tmp.go.txt"), strings.Join(srclns, "\n")); e != nil {
		panic(e)
	}
}

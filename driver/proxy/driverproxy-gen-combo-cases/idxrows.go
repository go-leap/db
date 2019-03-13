package main

import (
	"sort"
)

type idxRow []int

func (me idxRow) has(v int) bool {
	for i := range me {
		if me[i] == v {
			return true
		}
	}
	return false
}

func (me idxRow) eq(that idxRow) bool {
	if l := len(me); l == len(that) {
		for i := 0; i < l; i++ {
			if me[i] != that[i] {
				return false
			}
		}
		return true
	}
	return false
}

type idxRows []idxRow

func (me idxRows) Len() int           { return len(me) }
func (me idxRows) Swap(i, j int)      { me[i], me[j] = me[j], me[i] }
func (me idxRows) Less(i, j int) bool { return len(me[j]) < len(me[i]) }

func (me idxRows) has(sortedRow idxRow) bool {
	for i := range me {
		if me[i].eq(sortedRow) {
			return true
		}
	}
	return false
}

func getIdxRows(n int, max int) (rows idxRows) {
	if n == 1 {
		for i := 0; i < max; i++ {
			rows = append(rows, idxRow{i})
		}
	} else {
		rows = getIdxRows(n-1, max)
		for i, l := 0, len(rows); i < l; i++ {
			for j := 0; j < max; j++ {
				if srcrow := rows[i]; !srcrow.has(j) {
					row := append(idxRow{j}, srcrow...)
					if sort.Ints(row); !rows.has(row) {
						rows = append(rows, row)
					}
				}
			}
		}
	}
	if n == max {
		sort.Sort(rows)
	}
	return
}

package report

import (
	"fmt"
	"strings"
)

type Record struct {
	Id     int
	Domain string
	Total  int
	Error  int
	Valid  int
	All    int
}

func (r *Record) IncError() {
	r.Error++
	r.Total++
}
func (r *Record) IncValid() {
	r.Valid++
	r.Total++
}
func (r Record) String() string {
	if len(r.Domain) < 30 {
		r.Domain += strings.Repeat(" ", 30-len(r.Domain))
	}
	tot := fmt.Sprintf("[%d / %d]", r.Total, r.All)
	if len(tot) < 15 {
		tot += strings.Repeat(" ", 15-len(tot))
	}

	val := fmt.Sprintf("valid: %d", r.Valid)
	if len(val) < 13 {
		val += strings.Repeat(" ", 13-len(val))
	}
	er := fmt.Sprintf("invalid: %d", r.Error)
	if len(er) < 13 {
		er += strings.Repeat(" ", 13-len(er))
	}

	return fmt.Sprintf("%s : %s | %s | %s",
		r.Domain, tot, val, er)
}

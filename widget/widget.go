// This file contains a few routines for parsing form values
package widget

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/skypies/util/date"
)

// {{{ FormValueDateRange

// This widget assumes the values 'date', 'range_from', and 'range_to'
func FormValueDateRange(r *http.Request) (s,e time.Time, err error) {
	err = nil

	switch r.FormValue("date") {
	case "today":
		s,_ = date.WindowForToday()
		e=s
	case "yesterday":
		s,_ = date.WindowForYesterday()
		e=s
	case "range":
		s = date.ArbitraryDatestring2MidnightPdt(r.FormValue("range_from"), "2006/01/02")
		e = date.ArbitraryDatestring2MidnightPdt(r.FormValue("range_to"), "2006/01/02")
		if s.After(e) { s,e = e,s }
	}
	
	e = e.Add(23*time.Hour + 59*time.Minute + 59*time.Second) // make sure e covers its whole day

	return
}

// }}}
// {{{ FormValueInt64

func FormValueInt64(r *http.Request, name string) int64 {
	val,_ := strconv.ParseInt(r.FormValue(name), 10, 64)
	return val
}

// }}}
// {{{ FormValueEpochTime

func FormValueEpochTime(r *http.Request, name string) time.Time {
	return time.Unix(FormValueInt64(r,name), 0)
}

// }}}
// {{{ FormValueFloat64

func FormValueFloat64(w http.ResponseWriter, r *http.Request, name string) float64 {	
	if val,err := strconv.ParseFloat(r.FormValue(name), 64); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return -1
	} else {
		return val
	}
}

// }}}
// {{{ FormValueFloat64EatErrs

func FormValueFloat64EatErrs(r *http.Request, name string) float64 {	
	if val,err := strconv.ParseFloat(r.FormValue(name), 64); err != nil {
		return 0.0
	} else {
		return val
	}
}

// }}}
// {{{ FormValueCheckbox

func FormValueCheckbox(r *http.Request, name string) bool {
	if r.FormValue(name) != "" {
		return true
	}
	return false
}

// }}}
// {{{ FormValueCommaSepStrings

func FormValueCommaSepStrings(r *http.Request, name string) []string {
	ret := []string{}
	r.ParseForm()
	for _,v := range r.Form[name] {
		for _,str := range strings.Split(v, ",") {
			if str != "" {
				ret = append(ret, str)
			}
		}
	}
	return ret
}

// }}}

// {{{ DateRangeToCGIArgs

func DateRangeToCGIArgs(s,e time.Time) string {
	str := fmt.Sprintf("date=range&range_from=%s&range_to=%s",
		s.Format("2006/01/02"), e.Format("2006/01/02"))
	return str
}

// }}}


// {{{ -------------------------={ E N D }=----------------------------------

// Local variables:
// folded-file: t
// end:

// }}}

package histogram

import(
	"fmt"
	"sort"
	hdr "github.com/codahale/hdrhistogram"
)

/*

 s := histogram.NewSet()
 s.RecordValue("event_name", int64(millis))
 fmt.Printf("Stats:-\n%s", s)

 */

// {{{ Set{}

type Set struct {
	m map[string]*hdr.Histogram
}

// }}}

// {{{ NewSet

func NewSet() Set {
	return Set{
		m: map[string]*hdr.Histogram{},
	}
}

// }}}
// {{{ s.RecordValue

func (s Set)RecordValue(name string, val int64) {
	if _,exists := s.m[name]; !exists {
		s.m[name] = hdr.New(0,10000000, 2)
	}

	s.m[name].RecordValue(val)
}

// }}}
// {{{ s.String

func (s Set)String() string {
	names := []string{}
	for name,_ := range s.m { names = append(names,name) }
	sort.Strings(names)

	str := ""
	for _,name := range names {
		// Width is 2000; if using millis, that'll be two seconds, anything longer goes in overflow
		str += fmt.Sprintf("%-20.20s: %s\n", name, HDR2ASCII(s.m[name], 40, 0, 2000))
	}
	return str
}

// }}}


// {{{ -------------------------={ E N D }=----------------------------------

// Local variables:
// folded-file: t
// end:

// }}}

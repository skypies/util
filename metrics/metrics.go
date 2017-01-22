package metrics

import(
	"fmt"
	"sort"
	"time"
	hdr "github.com/codahale/hdrhistogram"
	histutil "github.com/skypies/util/histogram"
)

// {{{ Metric{}

type Metric struct {
	Name string
	wh *hdr.WindowedHistogram  // Holds the most recent 60 minutes
	snaps []*hdr.Snapshot       // Holds the previous 24 hours
}

func (m *Metric)String() string {
	str := fmt.Sprintf("%s\n  60s: %s\n  60m: %s\n",
		m.Name,
		histutil.HDR2ASCII(m.wh.Current, 40, 0, 1000),
		histutil.HDR2ASCII(m.wh.Merge(), 40, 0, 1000))

	for i,snap := range m.snaps {
		str += fmt.Sprintf(" +%02dh: %s\n", i, histutil.HDR2ASCII(hdr.Import(snap), 40, 0, 1000))
	}
	
	return str
}

func NewMetric(name string) *Metric {
	// 60 underlying 'grams; one per minute; cover an hour.
	// Covers values 0-10M (milliseconds, so 0-10000s).
	// Only 2 sig figs - but that's enough for us.
	return &Metric{Name:name, wh:hdr.NewWindowed(60, 0,10000000, 2), snaps:[]*hdr.Snapshot{}}
}

func (m *Metric)RecordValue(v int64) {
	m.wh.Current.RecordValue(v) // ignore error
}

func (m *Metric)Rotate() {
	m.wh.Rotate()
}

func (m *Metric)StoreSnapshot() {
	snap := m.wh.Merge().Export()
	m.snaps = append(m.snaps, snap)
	if len(m.snaps) > 24 {
		m.snaps = m.snaps[:24]
	}
}

// }}}


// {{{ Metrics{}

type Metrics struct {
	m map[string]*Metric

	// State to help decide when to rotate from minute to minute, and when to take a snap
	lastRotation   time.Time
	numRotations   int
}

// }}}

// {{{ NewMetrics

func NewMetrics() Metrics {
	return Metrics{
		m: map[string]*Metric{},
		lastRotation: time.Now(),
		numRotations: 0,
	}
}

// }}}
// {{{ m.RecordValue

func (m *Metrics)RecordValue(name string, val int64) {
	// Do we need to roll things ?
	minsSinceLastRotation := int(time.Since(m.lastRotation).Minutes())
	for i:=0; i<minsSinceLastRotation; i++ {
		m.lastRotation = time.Now()
		m.numRotations++

		for name,_ := range m.m {
			m.m[name].Rotate()
		}

		if m.numRotations >= 60 {
			for name,_ := range m.m {
				m.m[name].StoreSnapshot()
			}
			m.numRotations = 0
		}
	}

	if _,exists := m.m[name]; !exists {
		m.m[name] = NewMetric(name)
	}

	m.m[name].RecordValue(val)
}

// }}}
// {{{ m.String

func (m *Metrics)String() string {
	names := []string{}
	for name,_ := range m.m { names = append(names,name) }
	sort.Strings(names)

	str := ""
	for _,name := range names {
		str += fmt.Sprintf("%s\n", m.m[name])
	}
	return str
}

// }}}

// {{{ -------------------------={ E N D }=----------------------------------

// Local variables:
// folded-file: t
// end:

// }}}

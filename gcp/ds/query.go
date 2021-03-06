package ds

import "fmt"

// Query is a thin skin over the datastore query API. It also provides a textual dump of the
// query.
type Query struct {
	Kind            string
	AncestorKeyer   Keyer
	Filters       []Filter
	ProjectFields []string
	OrderStr        string
	LimitVal        int
	KeysOnlyVal     bool
	DistinctVals    bool
}

type Filter struct {
	Field string
	Value interface{}
}

func (q *Query)String() string {
	str := fmt.Sprintf("NewQuery(%q)\n", q.Kind)
	if q.AncestorKeyer != nil { str += fmt.Sprintf("  .Ancestor(%v)\n", q.AncestorKeyer) }
	for _,f := range q.Filters {
		str += fmt.Sprintf("  .Filter(%q, %v)\n", f.Field, f.Value)
	}
	if len(q.ProjectFields) != 0 { str += fmt.Sprintf("  .Project%q\n", q.ProjectFields) }
	if q.OrderStr != ""          { str += fmt.Sprintf("  .Order(%q)\n", q.OrderStr) }
	if q.LimitVal != 0           { str += fmt.Sprintf("  .Limit(%d)\n", q.LimitVal) }
	if q.KeysOnlyVal             { str += fmt.Sprintf("  .KeysOnly()\n") }
	if q.DistinctVals            { str += fmt.Sprintf("  .Distinct()\n") }
	return str
}

func NewQuery(kind string) *Query { return &Query{Kind:kind} }

func (q *Query)Filter(field string, val interface{}) *Query {
	q.Filters = append(q.Filters, Filter{field, val})
	return q
}

func (q *Query)Project(fields ...string) *Query {
	q.ProjectFields = fields
	return q
}

func (q *Query)Order(o string) *Query {
	q.OrderStr = o
	return q
}

func (q *Query)Limit(l int) *Query {
	q.LimitVal = l
	return q
}

func (q *Query)KeysOnly() *Query {
	q.KeysOnlyVal = true
	return q
}

func (q *Query)Distinct() *Query {
	q.DistinctVals = true
	return q
}

func (q *Query)Ancestor(keyer Keyer) *Query {
	q.AncestorKeyer = keyer
	return q
}

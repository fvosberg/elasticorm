package elasticorm

type Query struct{}

func (q Query) String() string {
	if q.QueryContext() {
	} else {
	}
	return ""
}

// QueryContext returns a boolean indicating wether the resulting query should be for query context (with a calculated score) - true -
// or for filtering context (hard matches) - false -
func (q Query) QueryContext() bool {
	return false
}

package elasticorm

func NewTestSuite(url string) *TestSuite {
	return &TestSuite{url: url}
}

type TestSuite struct {
	url       string
	Existing  []interface{}
	Query     Query
	Expecting []interface{}
}

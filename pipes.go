package query

type Pipe struct {
	Id string
	// if it has an error ... live goes on
	Executor func(QueryResult) (bool, QueryResult)
	IsFilter bool
}

package query

type QueryError string

func (e QueryError) Error() string { return string(e) }

const ErrorMissingResource = QueryError("Missing resource in that query")
const ErrorMissingTarget = QueryError("Missing target in the query")
const ErrorNoExecutor = QueryError("A executor for that resource does not exist")
const ErrorAttrNotFound = QueryError("That attribute does not exist")
const ErrorNoAttrSelected = QueryError("No attribute was selected")
const ErrorMissingSelector = QueryError("Missing a selector")
const ErrorInvalidSelector = QueryError("Invalid selector")

package apperror

import "fmt"

type Err struct {
	StartErr error
	Message  string
	Code     int
}

func (e Err) Error() string {
	return fmt.Sprintf("%s failed: %s - %s", e.Error, e.Message, e.Code)
}

func QueryPrepareError(err error) error {
	return Err{err, "Unable to prepare query", 500}
}

func QueryError(err error) error {
	return Err{err, "Unable to query database", 500}
}

func QueryScanError(err error) error {
	return Err{err, "Unable to scan query result", 500}
}

func QueryStatementError(err error) error {
	return Err{err, "Unable to close database statement", 500}
}

package gomo

import "fmt"

const MsgGomongoConnectionError = "failed to connect to database"
const MsgGomongoCursorError = "failed to open cursor to query result"
const MsgGomongoFailedFindError = "failed to query database"
const MsgGomongoUnmarshalError = "failed to unmarshal document"
const MsgGomongoInsertManyError = "failed to insert many documents"
const MsgGomongoFetchError = "failed to fetch documents from database"
const MsgGomongoDeleteError = "failed to delete documents"
const MsgGomongoCommandError = "failed to run command"
const MsgGomongoIndexError = "index command failed"

type GomongoError struct {
	Err      error
	mongoErr error
}

func (ge *GomongoError) Error() string {
	return ge.Err.Error()
}

func (ge *GomongoError) Unwrap() error {
	return ge.mongoErr
}

func NewError(msg string, mongoerror error) *GomongoError {
	return &GomongoError{
		Err:      fmt.Errorf("%s %s", msg, mongoerror),
		mongoErr: mongoerror,
	}
}

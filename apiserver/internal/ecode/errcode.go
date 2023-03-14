package ecode

import "fmt"

// ErrCode 返回错误码和消息的结构体
type ErrCode struct {
	Code int
	Msg  string
}

func (code ErrCode) Error() string {
	return code.Msg
}

// Err represents an error
type Err struct {
	Code int
	Msg  string
	Err  error
}

func (err *Err) Error() string {
	return fmt.Sprintf("Err - code: %d, message: %s, error: %s", err.Code, err.Msg, err.Err)
}

// DecodeErr 对错误进行解码，返回错误code和错误提示
func DecodeErr(err error) (int, string) {
	if err == nil {
		return Success.Code, Success.Msg
	}

	switch typed := err.(type) {
	case *Err:
		return typed.Code, typed.Msg
	case *ErrCode:
		return typed.Code, typed.Msg
	default:
	}

	//return InternalServerError.Code, InternalServerError.Msg
	//return InternalServerError.Code, err.Error()
	return -1, err.Error()

}

package gerr

import (
	"github.com/pkg/errors"
	"strings"
)

var DefaultCodeRelation = newCodeRelation()

type CodeError interface {
	Code() int
	Msg() string
	Detail() string
	WithDetail(detail string) CodeError
	RequestID() string
	Error
}

// NewLang 返回多语言的错误信息 strings[0] 为language，strings[1] 为request_id
func NewLang(code int, strings ...string) CodeError {
	if len(strings) == 0 {
		return &codeError{
			code: code,
			msg:  GetLanguageMsg(code, DefaultLanguage),
		}
	} else if len(strings) == 1 {
		return &codeError{
			code: code,
			msg:  GetLanguageMsg(code, strings[0]),
		}
	} else {
		return &codeError{
			code:      code,
			msg:       GetLanguageMsg(code, strings[0]),
			requestID: strings[1],
		}
	}
}

// NewCode 自定义返回错误信息，不带request_id，可通过withRequestID方法添加request_id
func NewCode(code int, msg string) CodeError {
	return &codeError{
		code: code,
		msg:  msg,
	}
}

type codeError struct {
	code      int
	msg       string
	detail    string
	requestID string
}

func (e *codeError) Code() int {
	return e.code
}

func (e *codeError) Msg() string {
	return e.msg
}

func (e *codeError) Detail() string {
	return e.detail
}

func (e *codeError) RequestID() string {
	return e.requestID
}

func (e *codeError) WithDetail(detail string) CodeError {
	var d string
	if e.detail == "" {
		d = detail
	} else {
		d = e.detail + "," + detail
	}
	return &codeError{
		code:   e.code,
		msg:    e.msg,
		detail: d,
	}
}

func (e *codeError) WithRequestID(requestID string) CodeError {
	return &codeError{
		code:      e.code,
		msg:       e.msg,
		detail:    e.detail,
		requestID: requestID,
	}
}

func (e *codeError) Wrap() error {
	return Wrap(e)
}

func (e *codeError) WrapMsg(msg string, kv ...any) error {
	return WrapMsg(e, msg, kv...)
}

func (e *codeError) Is(err error) bool {
	var codeErr CodeError
	ok := errors.As(Unwrap(err), &codeErr)
	if !ok {
		if err == nil && e == nil {
			return true
		}
		return false
	}
	if e == nil {
		return false
	}
	code := codeErr.Code()
	if e.code == code {
		return true
	}
	return DefaultCodeRelation.Is(e.code, code)
}

const initialCapacity = 3

func (e *codeError) Error() string {
	v := make([]string, 0, initialCapacity)
	v = append(v, e.msg)

	if e.detail != "" {
		v = append(v, e.detail)
	}

	return strings.Join(v, "|")
}

func Unwrap(err error) error {
	for err != nil {
		unwrap, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		err = unwrap.Unwrap()
	}
	return err
}

func Wrap(err error) error {
	return errors.WithStack(err)
}

func WrapMsg(err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	withMessage := errors.WithMessage(err, toString(msg, kv))
	return errors.WithStack(withMessage)
}

type CodeRelation interface {
	Add(codes ...int) error
	Is(parent, child int) bool
}

func newCodeRelation() CodeRelation {
	return &codeRelation{m: make(map[int]map[int]struct{})}
}

type codeRelation struct {
	m map[int]map[int]struct{}
}

const minimumCodesLength = 2

func (r *codeRelation) Add(codes ...int) error {
	if len(codes) < minimumCodesLength {
		return New("codes length must be greater than 2", "codes", codes).Wrap()
	}
	for i := 1; i < len(codes); i++ {
		parent := codes[i-1]
		s, ok := r.m[parent]
		if !ok {
			s = make(map[int]struct{})
			r.m[parent] = s
		}
		for _, code := range codes[i:] {
			s[code] = struct{}{}
		}
	}
	return nil
}

func (r *codeRelation) Is(parent, child int) bool {
	if parent == child {
		return true
	}
	s, ok := r.m[parent]
	if !ok {
		return false
	}
	_, ok = s[child]
	return ok
}

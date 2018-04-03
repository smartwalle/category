package payment

import "errors"

var (
	ErrSubjectNotAllowed = errors.New("subject 不能为空字符")
)

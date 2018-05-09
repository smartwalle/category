package category

import "errors"

var (
	ErrCategoryNotExists       = errors.New("分类不存在")
	ErrParentCategoryNotExists = errors.New("父分类不存在")
	ErrParentNotAllowed        = errors.New("不满足条件的父分类")
	ErrUnknownPosition         = errors.New("未知位置")
)

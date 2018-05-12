package category

import "github.com/smartwalle/dbs"

type Service struct {
	m *manager
}

func NewService(db dbs.DB, table string) *Service {
	var s = &Service{}
	var m = &manager{}
	m.db = db
	m.table = table
	s.m = m
	return s
}

// --------------------------------------------------------------------------------
// AddRoot 添加顶级分类
func (this *Service) AddRoot(cType int, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(0, cType, k_ADD_CATEGORY_POSITION_ROOT, 0, name, description, status, ext...)
}

func (this *Service) AddRootWithId(cId int64, cType int, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(cId, cType, k_ADD_CATEGORY_POSITION_ROOT, 0, name, description, status, ext...)
}

// AddToFirst 添加子分类，新添加的子分类位于子分类列表的前面
func (this *Service) AddToFirst(referTo int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(0, -1, k_ADD_CATEGORY_POSITION_FIRST, referTo, name, description, status, ext...)
}

func (this *Service) AddToFirstWithId(referTo, cId int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(cId, -1, k_ADD_CATEGORY_POSITION_FIRST, referTo, name, description, status, ext...)
}

// AddToLast 添加子分类，新添加的子分类位于子分类列表的后面
func (this *Service) AddToLast(referTo int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(0, -1, k_ADD_CATEGORY_POSITION_LAST, referTo, name, description, status, ext...)
}

func (this *Service) AddToLastWithId(referTo, cId int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(cId, -1, k_ADD_CATEGORY_POSITION_LAST, referTo, name, description, status, ext...)
}

// AddToLeft 添加兄弟分类，新添加的分类位于指定分类的左边(前面)
func (this *Service) AddToLeft(referTo int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(0, -1, k_ADD_CATEGORY_POSITION_LEFT, referTo, name, description, status, ext...)
}

func (this *Service) AddToLeftWithId(referTo, cId int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(cId, -1, k_ADD_CATEGORY_POSITION_LEFT, referTo, name, description, status, ext...)
}

// AddToRight 添加兄弟分类，新添加的分类位于指定分类的右边(后面)
func (this *Service) AddToRight(referTo int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(0, -1, k_ADD_CATEGORY_POSITION_RIGHT, referTo, name, description, status, ext...)
}

func (this *Service) AddToRightWithId(referTo, cId int64, name, description string, status int, ext ...string) (result int64, err error) {
	return this.m.addCategory(cId, -1, k_ADD_CATEGORY_POSITION_RIGHT, referTo, name, description, status, ext...)
}

// --------------------------------------------------------------------------------
// GetCategory 获取分类信息
// id 分类 id
func (this *Service) GetCategory(id int64) (result *Category, err error) {
	return this.m.getCategory(id)
}

func (this *Service) GetCategoryWithName(cType int, name string) (result *Category, err error) {
	return this.m.getCategoryWithName(cType, name)
}

// GetCategoryAdvList 获取分类列表
// parentId: 父分类id，当此参数的值大于 0 的时候，将忽略 cType 参数
// cType: 指定筛选分类的类型
// status: 指定筛选分类的状态
// depth: 指定要获取多少级别内的分类
func (this *Service) GetCategoryAdvList(parentId int64, cType, status, depth int, name string, limit uint64) (result []*Category, err error) {
	return this.m.getCategoryList(parentId, cType, status, depth, name, limit, false)
}

// GetNodeList 获取指定分类的子分类
func (this *Service) GetNodeList(parentId int64, status, depth int) (result []*Category, err error) {
	return this.m.getCategoryList(parentId, 0, status, depth, "", 0, false)
}

// GetNodeIdList 获取指定分类的子分类 id 列表
func (this *Service) GetNodeIdList(parentId int64, status, depth int) (result []int64, err error) {
	return this.m.getIdList(parentId, status, depth, false)
}

// GetCategoryList 获取指定分类的子分类列表，返回的列表包含指定的分类
func (this *Service) GetCategoryList(parentId int64, status, depth int) (result []*Category, err error) {
	return this.m.getCategoryList(parentId, 0, status, depth, "", 0, true)
}

// GetIdList 获取指定分类的子分类 id 列表，返回的 id 列表包含指定的分类
func (this *Service) GetIdList(parentId int64, status, depth int) (result []int64, err error) {
	return this.m.getIdList(parentId, status, depth, true)
}

// GetParentList 获取指定分类的父分类列表
func (this *Service) GetParentList(id int64, status int) (result []*Category, err error) {
	return this.m.getPathList(id, status, false)
}

// GetPathList 获取指定分类到 root 分类的完整分类列表，包括自身
func (this *Service) GetPathList(id int64, status int) (result []*Category, err error) {
	return this.m.getPathList(id, status, true)
}

// --------------------------------------------------------------------------------
// UpdateCategory 更新分类信息
func (this *Service) UpdateCategory(id int64, name, description, ext1, ext2 string) (err error) {
	return this.m.updateCategory(id, name, description, ext1, ext2)
}

// updateCategoryStatus 更新分类状态
// id: 被更新分类的 id
// status: 新的状态
// updateType:
// 		0、只更新当前分类的状态，子分类的状态不会受到影响，并且不会改变父子关系；
// 		1、子分类的状态会一起更新，不会改变父子关系；
// 		2、子分类的状态不会受到影响，并且所有子分类会向上移动一级（只针对把状态设置为 无效 的时候）；
func (this *Service) UpdateCategoryStatus(id int64, status, updateType int) (err error) {
	return this.m.updateCategoryStatus(id, status, updateType)
}

func (this *Service) MoveToRoot(id int64) (err error) {
	return this.m.moveCategory(k_MOVE_CATEGORY_POSITION_ROOT, id, 0)
}

func (this *Service) MoveToFirst(id, pid int64) (err error) {
	return this.m.moveCategory(k_MOVE_CATEGORY_POSITION_FIRST, id, pid)
}

func (this *Service) MoveToLast(id, pid int64) (err error) {
	return this.m.moveCategory(k_MOVE_CATEGORY_POSITION_LAST, id, pid)
}

func (this *Service) MoveToLeft(id, rid int64) (err error) {
	return this.m.moveCategory(k_MOVE_CATEGORY_POSITION_LEFT, id, rid)
}

func (this *Service) MoveToRight(id, rid int64) (err error) {
	return this.m.moveCategory(k_MOVE_CATEGORY_POSITION_RIGHT, id, rid)
}

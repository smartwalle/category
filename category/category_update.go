package category

import "github.com/smartwalle/dbs"

// UpdateCategory 更新分类信息
func (this *Manager) UpdateCategory(id int64, name, description string) (err error) {
	var ub = dbs.NewUpdateBuilder()
	ub.Table(this.table)
	ub.SET("name", name)
	ub.SET("description", description)
	ub.Where("id = ?", id)
	ub.Limit(1)
	if _, err = ub.Exec(this.db); err != nil {
		return nil
	}
	return nil
}

// UpdateCategoryStatus 更新分类状态
// id: 被更新分类的 id
// status: 新的状态
// updateType:
// 		0、只更新当前分类的状态，子分类的状态不会受到影响，并且不会改变父子关系；
// 		1、子分类的状态会一起更新，不会改变父子关系；
// 		2、子分类的状态不会受到影响，并且所有子分类会向上移动一级；
func (this *Manager) UpdateCategoryStatus(id int64, status, updateType int) (err error) {
	return nil
}
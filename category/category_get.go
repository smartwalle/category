package category

import (
	"github.com/smartwalle/dbs"
)

// --------------------------------------------------------------------------------
// GetCategory 获取分类信息
// id 分类 id
func (this *Manager) GetCategory(id int64) (result *Category, err error) {
	var tx = dbs.MustTx(this.db)

	if result, err = this.getCategoryWithId(tx, id); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *Manager) getCategoryWithId(tx *dbs.Tx, id int64) (result *Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("c.id", "c.type", "c.name", "c.description", "c.left_value", "c.right_value", "c.depth", "c.status", "c.ext1", "c.ext2", "c.created_on", "c.updated_on")
	sb.From(this.table, "AS c")
	sb.Where("c.id = ?", id)
	sb.Limit(1)
	if err = tx.ExecSelectBuilder(sb, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *Manager) getCategoryWithMaxRightValue(tx *dbs.Tx, cType int) (result *Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("c.id", "c.type", "c.name", "c.description", "c.left_value", "c.right_value", "c.depth", "c.status", "c.ext1", "c.ext2", "c.created_on", "c.updated_on")
	sb.From(this.table, "AS c")
	if cType > 0 {
		sb.Where("c.type = ?", cType)
	}
	sb.OrderBy("c.right_value DESC")
	sb.Limit(1)
	if err = tx.ExecSelectBuilder(sb, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetCategoryAdvList 获取分类列表
// parentId: 父分类id，当此参数的值大于 0 的时候，将忽略 cType 参数
// cType: 指定筛选分类的类型
// status: 指定筛选分类的状态
// depth: 指定要获取多少级别内的分类
func (this *Manager) GetCategoryAdvList(parentId int64, cType, status, depth int, name string, limit uint64) (result []*Category, err error) {
	return this.getCategoryList(parentId, cType, status, depth, name, limit, false)
}

func (this *Manager) getCategoryList(parentId int64, cType, status, depth int, name string, limit uint64, includeParent bool) (result []*Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("c.id", "c.type", "c.name", "c.description", "c.left_value", "c.right_value", "c.depth", "c.status", "c.ext1", "c.ext2", "c.created_on", "c.updated_on")
	sb.From(this.table, "AS c")
	if parentId > 0 {
		if includeParent {
			sb.LeftJoin(this.table, "AS pc ON pc.type = c.type AND pc.left_value <= c.left_value AND pc.right_value >= c.right_value")
		} else {
			sb.LeftJoin(this.table, "AS pc ON pc.type = c.type AND pc.left_value < c.left_value AND pc.right_value > c.right_value")
		}
		sb.Where("pc.id = ?", parentId)
	} else {
		if cType > 0 {
			sb.Where("c.type = ?", cType)
		}
	}
	if status > 0 {
		sb.Where("c.status = ?", status)
	}
	if depth > 0 {
		if parentId > 0 {
			sb.Where("c.depth - pc.depth <= ?", depth)
		} else {
			sb.Where("c.depth <= ?", depth)
		}
	}
	if name != "" {
		sb.Where("c.name = ?", name)
	}
	sb.OrderBy("c.type", "c.left_value")
	if limit > 0 {
		sb.Limit(limit)
	}

	err = sb.Scan(this.db, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetNodeList 获取指定分类的子分类
func (this *Manager) GetNodeList(parentId int64, status, depth int) (result []*Category, err error) {
	return this.getCategoryList(parentId, 0, status, depth, "", 0, false)
}

// GetNodeIdList 获取指定分类的子分类 id 列表
func (this *Manager) GetNodeIdList(parentId int64, status, depth int) (result []int64, err error) {
	return this.getIdList(parentId, status, depth, false)
}

// GetCategoryList 获取指定分类的子分类列表，返回的列表包含指定的分类
func (this *Manager) GetCategoryList(parentId int64, status, depth int) (result []*Category, err error) {
	return this.getCategoryList(parentId, 0, status, depth, "", 0, true)
}

// GetIdList 获取指定分类的子分类 id 列表，返回的 id 列表包含指定的分类
func (this *Manager) GetIdList(parentId int64, status, depth int) (result []int64, err error) {
	return this.getIdList(parentId, status, depth, true)
}

func (this *Manager) getIdList(parentId int64, status, depth int, includeParent bool) (result []int64, err error) {
	categoryList, err := this.getCategoryList(parentId, 0, status, depth, "", 0, includeParent)
	if err != nil {
		return nil, err
	}
	for _, c := range categoryList {
		result = append(result, c.Id)
	}
	return result, nil
}

// GetParentList 获取指定分类的父分类列表
func (this *Manager) GetParentList(id int64, status int) (result []*Category, err error) {
	return this.getPathList(id, status, false)
}

// GetPathList 获取指定分类到 root 分类的完整分类列表，包括自身
func (this *Manager) GetPathList(id int64, status int) (result []*Category, err error) {
	return this.getPathList(id, status, true)
}

func (this *Manager) getPathList(id int64, status int, includeLastNode bool) (result []*Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("pc.id", "pc.type", "pc.name", "pc.description", "pc.left_value", "pc.right_value", "pc.depth", "pc.status", "pc.ext1", "pc.ext2", "pc.created_on", "pc.updated_on")
	sb.From(this.table, "AS c")
	if includeLastNode {
		sb.LeftJoin(this.table, "AS pc ON pc.type = c.type AND pc.left_value <= c.left_value AND pc.right_value >= c.right_value")
	} else {
		sb.LeftJoin(this.table, "AS pc ON pc.type = c.type AND pc.left_value < c.left_value AND pc.right_value > c.right_value")
	}
	sb.Where("c.id = ?", id)
	if status > 0 {
		sb.Where("pc.status = ?", status)
	}
	sb.OrderBy("pc.left_value")
	err = sb.Scan(this.db, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

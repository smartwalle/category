package category

import "github.com/smartwalle/dbs"

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
	sb.Selects("id", "type", "name", "description", "left_value", "right_value", "status", "created_on", "updated_on")
	sb.From(this.table)
	sb.Where("id=?", id)
	sb.Limit(1)
	if err = tx.ExecSelectBuilder(sb, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *Manager) getCategoryWithMaxRightValue(tx *dbs.Tx, cType int) (result *Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("id", "type", "name", "description", "left_value", "right_value", "status", "created_on", "updated_on")
	sb.From(this.table)
	if cType > 0 {
		sb.Where("type =? ", cType)
	}
	sb.OrderBy("right_value DESC")
	sb.Limit(1)
	if err = tx.ExecSelectBuilder(sb, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *Manager) GetCategoryList(parentId int64, cType, status int) (result []*Category, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("c.id", "c.type", "c.name", "c.description", "c.left_value", "c.right_value", "c.status", "c.created_on", "c.updated_on")
	sb.From(this.table, "AS c")
	if parentId > 0 {
		sb.LeftJoin(this.table, "AS pc ON pc.left_value <= c.left_value AND pc.right_value >= c.right_value")
		sb.Where("pc.id = ?", parentId)
	}
	if cType > 0 {
		sb.Where("c.type = ?", cType)
	}
	if status > 0 {
		sb.Where("c.status = ?", status)
	}
	sb.OrderBy("c.type")
	sb.OrderBy("c.left_value")
	err = sb.Scan(this.db, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

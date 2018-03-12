package category

import (
	"github.com/smartwalle/dbs"
	"time"
)

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
// 		2、子分类的状态不会受到影响，并且所有子分类会向上移动一级（只针对把状态设置为 无效 的时候）；
func (this *Manager) UpdateCategoryStatus(id int64, status, updateType int) (err error) {
	var sess = this.db

	// 锁表
	var lock = dbs.WriteLock(this.table)
	if _, err = lock.Exec(sess); err != nil {
		return err
	}

	// 解锁
	defer func() {
		var unlock = dbs.UnlockTable()
		unlock.Exec(sess)
	}()

	var tx = dbs.MustTx(sess)

	category, err := this.getCategoryWithId(tx, id)
	if err != nil {
		return err
	}

	if category == nil {
		tx.Rollback()
		return ErrCategoryNotExists
	}

	if category.Status == status {
		tx.Rollback()
		return nil
	}

	switch updateType {
	case 2:
		if status == K_CATEGORY_STATUS_DISABLE {
			var ub = dbs.NewUpdateBuilder()
			ub.Table(this.table)
			ub.SET("status", status)
			ub.SET("right_value", dbs.SQL("left_value+1"))
			ub.SET("updated_on", time.Now())
			ub.Where("id = ?", id)
			ub.Limit(1)
			if _, err := tx.ExecUpdateBuilder(ub); err != nil {
				return err
			}

			var ubChild = dbs.NewUpdateBuilder()
			ubChild.Table(this.table)
			ubChild.SET("left_value", dbs.SQL("left_value+1"))
			ubChild.SET("right_value", dbs.SQL("right_value+1"))
			ubChild.SET("depth", dbs.SQL("depth-1"))
			ubChild.SET("updated_on", time.Now())
			ubChild.Where("type = ? AND left_value > ? AND right_value < ?", category.Type, category.LeftValue, category.RightValue)
			if _, err := tx.ExecUpdateBuilder(ubChild); err != nil {
				return err
			}
		}
	case 1:
		var ub = dbs.NewUpdateBuilder()
		ub.Table(this.table)
		ub.SET("status", status)
		ub.SET("updated_on", time.Now())
		ub.Where("type = ? AND left_value >= ? AND right_value <= ?", category.Type, category.LeftValue, category.RightValue)
		if _, err := tx.ExecUpdateBuilder(ub); err != nil {
			return err
		}
	case 0:
		var ub = dbs.NewUpdateBuilder()
		ub.Table(this.table)
		ub.SET("status", status)
		ub.SET("updated_on", time.Now())
		ub.Where("id = ?", id)
		ub.Limit(1)
		if _, err := tx.ExecUpdateBuilder(ub); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (this *Manager) UpdateCategoryParent(id, pid int64) (err error) {
	if id == pid {
		return ErrParentNotAllowed
	}

	var sess = this.db

	// 锁表
	var lock = dbs.WriteLock(this.table)
	if _, err = lock.Exec(sess); err != nil {
		return err
	}

	// 解锁
	defer func() {
		var unlock = dbs.UnlockTable()
		unlock.Exec(sess)
	}()

	var tx = dbs.MustTx(sess)

	// 判断被移动的分类是否存在
	category, err := this.getCategoryWithId(tx, id)
	if err != nil {
		return err
	}
	if category == nil {
		tx.Rollback()
		return ErrCategoryNotExists
	}

	// 判断目标父分类是否存在
	parent, err := this.getCategoryWithId(tx, pid)
	if err != nil {
		return err
	}
	if parent == nil {
		tx.Rollback()
		return ErrParentCategoryNotExists
	}

	// 判断被移动分类和目标父分类是否属于同一 type
	if parent.Type != category.Type {
		tx.Rollback()
		return ErrParentNotAllowed
	}

	// 循环连接问题，即 目标父分类 是 被移动分类 的子分类
	if parent.LeftValue > category.LeftValue && parent.RightValue < category.RightValue {
		tx.Rollback()
		return ErrParentNotAllowed
	}

	// 判断是否已经是子分类
	if parent.LeftValue < category.LeftValue && parent.RightValue > category.RightValue && category.Depth - 1 == parent.Depth {
		tx.Rollback()
		return ErrParentNotAllowed
	}

	// 查询出被移动分类的所有子分类
	children, err := this.GetCategoryList(category.Id, 0, 0, 0, "", 0)
	if err != nil {
		return err
	}

	var updateIdList []int64
	updateIdList = append(updateIdList, category.Id)
	for _, c := range children {
		updateIdList = append(updateIdList, c.Id)
	}

	var diff = category.RightValue - category.LeftValue + 1

	// 把要移动的节点及其子节点从原树中删除掉
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", category.Type, category.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}
	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value > ?", category.Type, category.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	//if parent.RightValue > category.RightValue {
	//	if parent, err = this.getCategoryWithId(tx, parent.Id); err != nil {
	//		return err
	//	}
	//}
	if parent.LeftValue > category.RightValue {
		parent.LeftValue -= diff
	}
	if parent.RightValue > category.RightValue {
		parent.RightValue -= diff
	}

	// 移出空间用于存放被移动的节点及其子节点
	ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value + ?", diff))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", parent.Type, parent.RightValue)
	ubTreeLeft.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}

	ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value + ?", diff))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value >= ?", parent.Type, parent.RightValue)
	ubTreeRight.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	//if parent, err = this.getCategoryWithId(tx, parent.Id); err != nil {
	//	return err
	//}
	parent.RightValue += diff

	// 更新被移动节点的信息
	diff = category.RightValue - parent.RightValue + 1
	var diffDepth = parent.Depth - category.Depth + 1
	var ubTree = dbs.NewUpdateBuilder()
	ubTree.Table(this.table)
	ubTree.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTree.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTree.SET("depth", dbs.SQL("depth + ?",  diffDepth))
	ubTree.SET("updated_on", time.Now())
	ubTree.Where(dbs.IN("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTree); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
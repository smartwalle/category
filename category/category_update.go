package category

import (
	"github.com/smartwalle/dbs"
	"time"
)

const (
	k_MOVE_CATEGORY_POSITION_ROOT  = 0 // 顶级分类
	k_MOVE_CATEGORY_POSITION_FIRST = 1 // 列表头部 (子分类)
	k_MOVE_CATEGORY_POSITION_LAST  = 2 // 列表尾部 (子分类)
	k_MOVE_CATEGORY_POSITION_LEFT  = 3 // 左边 (兄弟分类)
	k_MOVE_CATEGORY_POSITION_RIGHT = 4 // 右边 (兄弟分类)
)

func (this *manager) updateCategory(id int64, name, description, ext1, ext2 string) (err error) {
	var ub = dbs.NewUpdateBuilder()
	ub.Table(this.table)
	ub.SET("name", name)
	ub.SET("description", description)
	ub.SET("ext1", ext1)
	ub.SET("ext2", ext2)
	ub.Where("id = ?", id)
	ub.Limit(1)
	if _, err = ub.Exec(this.db); err != nil {
		return nil
	}
	return nil
}

// updateCategoryStatus 更新分类状态
// id: 被更新分类的 id
// status: 新的状态
// updateType:
// 		0、只更新当前分类的状态，子分类的状态不会受到影响，并且不会改变父子关系；
// 		1、子分类的状态会一起更新，不会改变父子关系；
// 		2、子分类的状态不会受到影响，并且所有子分类会向上移动一级（只针对把状态设置为 无效 的时候）；
func (this *manager) updateCategoryStatus(id int64, status, updateType int) (err error) {
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
			ub.SET("right_value", dbs.SQL("left_value + 1"))
			ub.SET("updated_on", time.Now())
			ub.Where("id = ?", id)
			ub.Limit(1)
			if _, err := tx.ExecUpdateBuilder(ub); err != nil {
				return err
			}

			var ubChild = dbs.NewUpdateBuilder()
			ubChild.Table(this.table)
			ubChild.SET("left_value", dbs.SQL("left_value + 1"))
			ubChild.SET("right_value", dbs.SQL("right_value + 1"))
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

func (this *manager) moveCategory(position int, id, rid int64) (err error) {
	if id == rid {
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

	// 判断参照分类是否存在
	var referCategory *Category
	if position == k_MOVE_CATEGORY_POSITION_ROOT {
		// 如果是添加顶级分类，那么参照分类为 right value 最大的
		if referCategory, err = this.getCategoryWithMaxRightValue(tx, category.Type); err != nil {
			return err
		}
		if referCategory != nil && referCategory.Id == category.Id {
			tx.Rollback()
			return nil
		}
	} else {
		if referCategory, err = this.getCategoryWithId(tx, rid); err != nil {
			return err
		}
	}
	if referCategory == nil {
		tx.Rollback()
		return ErrParentCategoryNotExists
	}

	// 判断被移动分类和目标参照分类是否属于同一 type
	if referCategory.Type != category.Type {
		tx.Rollback()
		return ErrParentNotAllowed
	}

	// 循环连接问题，即 参照分类 是 被移动分类 的子分类
	if referCategory.LeftValue > category.LeftValue && referCategory.RightValue < category.RightValue {
		tx.Rollback()
		return ErrParentNotAllowed
	}

	// 判断是否已经是子分类
	//if referCategory.LeftValue < category.LeftValue && referCategory.RightValue > category.RightValue && category.Depth - 1 == referCategory.Depth {
	//	tx.Rollback()
	//	return ErrParentNotAllowed
	//}

	// 查询出被移动分类的所有子分类
	children, err := this.GetCategoryList(category.Id, 0, 0)
	if err != nil {
		return err
	}

	var updateIdList []int64
	updateIdList = append(updateIdList, category.Id)
	for _, c := range children {
		updateIdList = append(updateIdList, c.Id)
	}

	if err = this.moveCategoryWithPosition(tx, position, category, referCategory, updateIdList); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (this *manager) moveCategoryWithPosition(tx *dbs.Tx, position int, category, referCategory *Category, updateIdList []int64) (err error) {
	var nodeLen = category.RightValue - category.LeftValue + 1

	// 把要移动的节点及其子节点从原树中删除掉
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value - ?", nodeLen))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", category.Type, category.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}
	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value - ?", nodeLen))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value > ?", category.Type, category.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	if referCategory.LeftValue > category.RightValue {
		referCategory.LeftValue -= nodeLen
	}
	if referCategory.RightValue > category.RightValue {
		referCategory.RightValue -= nodeLen
	}

	switch position {
	case k_MOVE_CATEGORY_POSITION_ROOT:
		return this.moveToRight(tx, category, referCategory, updateIdList, nodeLen)
	case k_MOVE_CATEGORY_POSITION_FIRST:
		return this.moveToFirst(tx, category, referCategory, updateIdList, nodeLen)
	case k_MOVE_CATEGORY_POSITION_LAST:
		return this.moveToLast(tx, category, referCategory, updateIdList, nodeLen)
	case k_MOVE_CATEGORY_POSITION_LEFT:
		return this.moveToLeft(tx, category, referCategory, updateIdList, nodeLen)
	case k_MOVE_CATEGORY_POSITION_RIGHT:
		return this.moveToRight(tx, category, referCategory, updateIdList, nodeLen)
	}
	tx.Rollback()
	return ErrUnknownPosition
}

func (this *manager) moveToFirst(tx *dbs.Tx, category, parent *Category, updateIdList []int64, nodeLen int) (err error) {
	// 移出空间用于存放被移动的节点及其子节点
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value + ?", nodeLen))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", parent.Type, parent.LeftValue)
	ubTreeLeft.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}

	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value + ?", nodeLen))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value > ?", parent.Type, parent.LeftValue)
	ubTreeRight.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	parent.RightValue += nodeLen

	// 更新被移动节点的信息
	var diff = category.LeftValue - parent.LeftValue - 1
	var diffDepth = parent.Depth - category.Depth + 1
	var ubTree = dbs.NewUpdateBuilder()
	ubTree.Table(this.table)
	ubTree.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTree.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTree.SET("depth", dbs.SQL("depth + ?", diffDepth))
	ubTree.SET("updated_on", time.Now())
	ubTree.Where(dbs.IN("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTree); err != nil {
		return err
	}

	return nil
}

func (this *manager) moveToLast(tx *dbs.Tx, category, parent *Category, updateIdList []int64, nodeLen int) (err error) {
	// 移出空间用于存放被移动的节点及其子节点
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value + ?", nodeLen))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", parent.Type, parent.RightValue)
	ubTreeLeft.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}

	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value + ?", nodeLen))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value >= ?", parent.Type, parent.RightValue)
	ubTreeRight.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	parent.RightValue += nodeLen

	// 更新被移动节点的信息
	var diff = category.RightValue - parent.RightValue + 1
	var diffDepth = parent.Depth - category.Depth + 1
	var ubTree = dbs.NewUpdateBuilder()
	ubTree.Table(this.table)
	ubTree.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTree.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTree.SET("depth", dbs.SQL("depth + ?", diffDepth))
	ubTree.SET("updated_on", time.Now())
	ubTree.Where(dbs.IN("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTree); err != nil {
		return err
	}

	return nil
}

func (this *manager) moveToLeft(tx *dbs.Tx, category, referCategory *Category, updateIdList []int64, nodeLen int) (err error) {
	// 移出空间用于存放被移动的节点及其子节点
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value + ?", nodeLen))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value >= ?", referCategory.Type, referCategory.LeftValue)
	ubTreeLeft.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}

	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value + ?", nodeLen))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value >= ?", referCategory.Type, referCategory.LeftValue)
	ubTreeRight.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	//referCategory.LeftValue += nodeLen

	// 更新被移动节点的信息
	var diff = category.LeftValue - referCategory.LeftValue
	var diffDepth = referCategory.Depth - category.Depth
	var ubTree = dbs.NewUpdateBuilder()
	ubTree.Table(this.table)
	ubTree.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTree.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTree.SET("depth", dbs.SQL("depth + ?", diffDepth))
	ubTree.SET("updated_on", time.Now())
	ubTree.Where(dbs.IN("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTree); err != nil {
		return err
	}

	return nil
}

func (this *manager) moveToRight(tx *dbs.Tx, category, referCategory *Category, updateIdList []int64, nodeLen int) (err error) {
	// 移出空间用于存放被移动的节点及其子节点
	var ubTreeLeft = dbs.NewUpdateBuilder()
	ubTreeLeft.Table(this.table)
	ubTreeLeft.SET("left_value", dbs.SQL("left_value + ?", nodeLen))
	ubTreeLeft.SET("updated_on", time.Now())
	ubTreeLeft.Where("type = ? AND left_value > ?", referCategory.Type, referCategory.RightValue)
	ubTreeLeft.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeLeft); err != nil {
		return err
	}

	var ubTreeRight = dbs.NewUpdateBuilder()
	ubTreeRight.Table(this.table)
	ubTreeRight.SET("right_value", dbs.SQL("right_value + ?", nodeLen))
	ubTreeRight.SET("updated_on", time.Now())
	ubTreeRight.Where("type = ? AND right_value > ?", referCategory.Type, referCategory.RightValue)
	ubTreeRight.Where(dbs.NotIn("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTreeRight); err != nil {
		return err
	}

	// 更新被移动节点的信息
	var diff = category.LeftValue - referCategory.RightValue - 1
	var diffDepth = referCategory.Depth - category.Depth
	var ubTree = dbs.NewUpdateBuilder()
	ubTree.Table(this.table)
	ubTree.SET("left_value", dbs.SQL("left_value - ?", diff))
	ubTree.SET("right_value", dbs.SQL("right_value - ?", diff))
	ubTree.SET("depth", dbs.SQL("depth + ?", diffDepth))
	ubTree.SET("updated_on", time.Now())
	ubTree.Where(dbs.IN("id", updateIdList))
	if _, err = tx.ExecUpdateBuilder(ubTree); err != nil {
		return err
	}

	return nil
}

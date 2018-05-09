package category

import (
	"github.com/smartwalle/dbs"
	"time"
)

const (
	k_ADD_CATEGORY_POSITION_ROOT  = 0 // 顶级分类
	k_ADD_CATEGORY_POSITION_FIRST = 1 // 列表头部 (子分类)
	k_ADD_CATEGORY_POSITION_LAST  = 2 // 列表尾部 (子分类)
	k_ADD_CATEGORY_POSITION_LEFT  = 3 // 左边 (兄弟分类)
	k_ADD_CATEGORY_POSITION_RIGHT = 4 // 右边 (兄弟分类)
)

// addCategory 添加分类
// cType: 分类类型（分类组）
// position:
// 		1、将新的分类添加到参照分类的子分类列表头部；
// 		2、将新的分类添加到参照分类的子分类列表尾部；
// 		3、将新的分类添加到参照分类的左边；
// 		4、将新的分类添加到参照分类的右边；
// referTo: 参照分类 id，如果值等于 0，则表示添加顶级分类
// name: 分类名
// description: 描述
// status: 分类状态 1000、有效；2000、无效
func (this *manager) addCategory(cId int64, cType, position int, referTo int64, name, description string, status int, ext ...string) (result *Category, err error) {
	var sess = this.db

	// 锁表
	var lock = dbs.WriteLock(this.table)
	if _, err = lock.Exec(sess); err != nil {
		return nil, err
	}

	// 解锁
	defer func() {
		var unlock = dbs.UnlockTable()
		unlock.Exec(sess)
	}()

	var tx = dbs.MustTx(sess)
	var newCategoryId int64 = 0

	// 查询出参照分类的信息
	var referCategory *Category

	if position == k_ADD_CATEGORY_POSITION_ROOT {
		// 如果是添加顶级分类，那么参照分类为 right value 最大的
		if referCategory, err = this.getCategoryWithMaxRightValue(tx, cType); err != nil {
			return nil, err
		}

		// 如果参照分类为 nil，则创建一个虚拟的
		if referCategory == nil {
			referCategory = &Category{}
			referCategory.Id = -1
			referCategory.Type = cType
			referCategory.LeftValue = 0
			referCategory.RightValue = 0
			referCategory.Depth = 1
		}
	} else {
		if referCategory, err = this.getCategoryWithId(tx, referTo); err != nil {
			return nil, err
		}
		if referCategory == nil {
			tx.Rollback()
			return nil, ErrCategoryNotExists
		}
	}

	var ext1 string
	var ext2 string
	if len(ext) > 0 {
		ext1 = ext[0]
	}
	if len(ext) > 1 {
		ext2 = ext[1]
	}

	if newCategoryId, err = this.addCategoryWithPosition(tx, referCategory, cId, position, name, description, ext1, ext2, status); err != nil {
		return nil, err
	}

	// 查询出刚刚添加的新分类
	if result, err = this.getCategoryWithId(tx, newCategoryId); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *manager) addCategoryWithPosition(tx *dbs.Tx, referCategory *Category,  cId int64, position int, name, description, ext1, ext2 string, status int) (id int64, err error) {
	switch position {
	case k_ADD_CATEGORY_POSITION_ROOT:
		return this.insertCategoryToRoot(tx, referCategory, cId, name, description, ext1, ext2, status)
	case k_ADD_CATEGORY_POSITION_FIRST:
		return this.insertCategoryToFirst(tx, referCategory, cId, name, description, ext1, ext2, status)
	case k_ADD_CATEGORY_POSITION_LAST:
		return this.insertCategoryToLast(tx, referCategory, cId, name, description, ext1, ext2, status)
	case k_ADD_CATEGORY_POSITION_LEFT:
		return this.insertCategoryToLeft(tx, referCategory, cId, name, description, ext1, ext2, status)
	case k_ADD_CATEGORY_POSITION_RIGHT:
		return this.insertCategoryToRight(tx, referCategory, cId, name, description, ext1, ext2, status)
	}
	tx.Rollback()
	return 0, ErrUnknownPosition
}

func (this *manager) insertCategoryToRoot(tx *dbs.Tx, referCategory *Category, cId int64, name, description, ext1, ext2 string, status int) (id int64, err error) {
	var cType = referCategory.Type
	var leftValue = referCategory.RightValue + 1
	var rightValue = referCategory.RightValue + 2
	var depth = referCategory.Depth
	if id, err = this.insertCategory(tx, cId, cType, name, description, ext1, ext2, leftValue, rightValue, depth, status); err != nil {
		return 0, err
	}
	return id, nil
}

func (this *manager) insertCategoryToFirst(tx *dbs.Tx, referCategory *Category, cId int64, name, description, ext1, ext2 string, status int) (id int64, err error) {
	var ubLeft = dbs.NewUpdateBuilder()
	ubLeft.Table(this.table)
	ubLeft.SET("left_value", dbs.SQL("left_value + 2"))
	ubLeft.SET("updated_on", time.Now())
	ubLeft.Where("type = ? AND left_value > ?", referCategory.Type, referCategory.LeftValue)
	if _, err = tx.ExecUpdateBuilder(ubLeft); err != nil {
		return 0, err
	}

	var ubRight = dbs.NewUpdateBuilder()
	ubRight.Table(this.table)
	ubRight.SET("right_value", dbs.SQL("right_value + 2"))
	ubRight.SET("updated_on", time.Now())
	ubRight.Where("type = ? AND right_value > ?", referCategory.Type, referCategory.LeftValue)
	if _, err = tx.ExecUpdateBuilder(ubRight); err != nil {
		return 0, err
	}

	if id, err = this.insertCategory(tx, cId, referCategory.Type, name, description, ext1, ext2, referCategory.LeftValue+1, referCategory.LeftValue+2, referCategory.Depth+1, status); err != nil {
		return 0, err
	}
	return id, nil
}

func (this *manager) insertCategoryToLast(tx *dbs.Tx, referCategory *Category, cId int64, name, description, ext1, ext2 string, status int) (id int64, err error) {
	var ubLeft = dbs.NewUpdateBuilder()
	ubLeft.Table(this.table)
	ubLeft.SET("left_value", dbs.SQL("left_value + 2"))
	ubLeft.SET("updated_on", time.Now())
	ubLeft.Where("type = ? AND left_value > ?", referCategory.Type, referCategory.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubLeft); err != nil {
		return 0, err
	}

	var ubRight = dbs.NewUpdateBuilder()
	ubRight.Table(this.table)
	ubRight.SET("right_value", dbs.SQL("right_value + 2"))
	ubRight.SET("updated_on", time.Now())
	ubRight.Where("type = ? AND right_value >= ?", referCategory.Type, referCategory.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubRight); err != nil {
		return 0, err
	}

	if id, err = this.insertCategory(tx, cId, referCategory.Type, name, description, ext1, ext2, referCategory.RightValue, referCategory.RightValue+1, referCategory.Depth+1, status); err != nil {
		return 0, err
	}

	return id, nil
}

func (this *manager) insertCategoryToLeft(tx *dbs.Tx, referCategory *Category, cId int64, name, description, ext1, ext2 string, status int) (id int64, err error) {
	var ubLeft = dbs.NewUpdateBuilder()
	ubLeft.Table(this.table)
	ubLeft.SET("left_value", dbs.SQL("left_value + 2"))
	ubLeft.SET("updated_on", time.Now())
	ubLeft.Where("type = ? AND left_value >= ?", referCategory.Type, referCategory.LeftValue)
	if _, err = tx.ExecUpdateBuilder(ubLeft); err != nil {
		return 0, err
	}

	var ubRight = dbs.NewUpdateBuilder()
	ubRight.Table(this.table)
	ubRight.SET("right_value", dbs.SQL("right_value + 2"))
	ubRight.SET("updated_on", time.Now())
	ubRight.Where("type = ? AND right_value >= ?", referCategory.Type, referCategory.LeftValue)
	if _, err = tx.ExecUpdateBuilder(ubRight); err != nil {
		return 0, err
	}

	if id, err = this.insertCategory(tx, cId, referCategory.Type, name, description, ext1, ext2, referCategory.LeftValue, referCategory.LeftValue+1, referCategory.Depth, status); err != nil {
		return 0, err
	}
	return id, nil
}

func (this *manager) insertCategoryToRight(tx *dbs.Tx, referCategory *Category, cId int64, name, description, ext1, ext2 string, status int) (id int64, err error) {
	var ubLeft = dbs.NewUpdateBuilder()
	ubLeft.Table(this.table)
	ubLeft.SET("left_value", dbs.SQL("left_value + 2"))
	ubLeft.SET("updated_on", time.Now())
	ubLeft.Where("type = ? AND left_value > ?", referCategory.Type, referCategory.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubLeft); err != nil {
		return 0, err
	}

	var ubRight = dbs.NewUpdateBuilder()
	ubRight.Table(this.table)
	ubRight.SET("right_value", dbs.SQL("right_value + 2"))
	ubRight.SET("updated_on", time.Now())
	ubRight.Where("type = ? AND right_value > ?", referCategory.Type, referCategory.RightValue)
	if _, err = tx.ExecUpdateBuilder(ubRight); err != nil {
		return 0, err
	}

	if id, err = this.insertCategory(tx, cId, referCategory.Type, name, description, ext1, ext2, referCategory.RightValue+1, referCategory.RightValue+2, referCategory.Depth, status); err != nil {
		return 0, err
	}
	return id, nil
}

func (this *manager) insertCategory(tx *dbs.Tx, cId int64, cType int, name, description, ext1, ext2 string, leftValue, rightValue, depth, status int) (id int64, err error) {
	var now = time.Now()
	var ib = dbs.NewInsertBuilder()
	ib.Table(this.table)
	ib.Columns("id", "type", "name", "description", "left_value", "right_value", "depth", "status", "ext1", "ext2", "created_on", "updated_on")
	ib.Values(cId, cType, name, description, leftValue, rightValue, depth, status, ext1, ext2, now, now)
	if result, err := tx.ExecInsertBuilder(ib); err != nil {
		return 0, err
	} else {
		id, _ = result.LastInsertId()
	}
	return id, err
}

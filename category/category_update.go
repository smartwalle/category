package category

import (
	"github.com/smartwalle/dbs"
	"github.com/smartwalle/errors"
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
	lock.WriteLock(this.table, "AS c")
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
		return errors.New("分类不存在")
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

## 左右值无限分类

本项目是采用 Go 实现的左右值无限分类，目前已经完成了主要接口的编写：

* 添加顶级分类
* 添加子分类（添加分类到指定分类的子分类列表的头部或者尾部）
* 添加兄弟分类（添加分类到指定分类的左边或者右边）
* 更新分类基本信息
* 更新分类的状态（更新状态的时候，可以选择是否影响子分类的状态）
* 改变分类的父分类
* 获取分类列表
* 获取分类的父分类列表

## 集成方法

### 创建表

首先创建如下结构的表

```sql
CREATE TABLE IF NOT EXISTS `category` (
	`id` int(11) NOT NULL AUTO_INCREMENT,
	`type` int(11) DEFAULT '0',
	`name` varchar(128) DEFAULT NULL,
	`description` varchar(512) DEFAULT NULL,
	`left_value` int(11) DEFAULT NULL,
	`right_value` int(11) DEFAULT NULL,
	`depth` int(11) DEFAULT NULL,
	`status` int(11) DEFAULT '1000',
	`created_on` datetime DEFAULT NULL,
	`updated_on` datetime DEFAULT NULL,
	PRIMARY KEY (`id`),
	UNIQUE KEY `category_id_uindex` (`id`),
	KEY `category_type_index` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;
```

### 项目集成

下载本代码库

```
go get github.com/smartwalle/m4go/
```

导入

```
import "github.com/smartwalle/dbs"
```

category.NewManager(...) 需要两个参数，一个是数据库连接对象，另一个是数据库中存放分类信息的表的名称

```go
// 创建数据库连接
var db, _ = sql.Open("mysql", "url")
var m = category.NewManager(db, "category")
m.GetCategoryList(...)
```

### 关于同表多业务

如果想要在同一个表中支持不同的业务分类，比如：商品分类、新闻分类、标签、话题分类等等，有两种实现方式：一是采用 type 进行区分；二是创建多个顶级分类，每一个顶级分类管理一个业务的分类。

在此推荐使用第一种方案，即采用 type 进行区分，能够有效提高数据更新的效率。
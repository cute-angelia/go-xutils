## usage


### import

```
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/cute-angelia/go-xutils/umysql"
)

```

### init

```
// 初始化-db
	gormopts1 := umysql.NewGormOpts(
		umysql.WithGormOptDbname("db_orange"),
		umysql.WithGormDsn("root:root@tcp(127.0.0.1:3306)/db_orange?charset=utf8&parseTime=true&loc=Local"),
		umysql.WithGormOptConnmax(3),
		umysql.WithGormLogDebug(true),
	)
	umysql.InitGorm(gormopts1)
```

### use 

```
orm := umysql.GetGorm("db_orange")

user := user.model{}

orm.Where("id = ?", 1).First(&user);
```
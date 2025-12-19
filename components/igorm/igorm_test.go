package igorm

import (
	"github.com/cute-angelia/go-xutils/components/loggers/loggerV3"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"gorm.io/gorm/logger"
	"log"
	"testing"
)

type Project struct {
	Id int32 `gorm:"primary_key;sort:desc" json:"id"` // Project

	Name string `gorm:"column:name" json:"name"` // Project

	Author string `gorm:"column:author" json:"author"` // Project

	LoginUid int32 `gorm:"column:login_uid" json:"login_uid"` // Project

	CreateAt string `gorm:"column:create_at" json:"create_at"` // Project

	UpdateAt string `gorm:"column:update_at" json:"update_at"` // Project

}

// TableName 设置表明
func (t Project) TableName() string {
	return "project"
}

func TestConnect(t *testing.T) {
	// loggerv3
	loggerV3.New()

	dbName := "company_ues"
	dsn := "admin"
	New(
		WithLogLevel(logger.Info),
		WithMaxIdleConns(1),
		WithMaxOpenConnss(5),
		WithDbName(dbName),
		WithDsn(dsn),
		WithLoggerWriter(loggerV3.GetLogger()),
	).MustInitMysql()

	orm, _ := GetGormMysql(dbName)
	model := Project{}
	orm.First(&model)
	t.Log(ijson.Pretty(model))

	orm2, _ := GetGormMysql(dbName)
	model2 := Project{} // 不新建数据有问题，和上面一样
	orm2.Debug().Where("1=1").Order("id desc").First(&model2)
	t.Log(ijson.Pretty(model2))
}

func TestQuery(t *testing.T) {
	t.Log("good")

	sql, v := NewQuery().
		Where("good1", "=", "hello").
		Where("good2", "=", 2).
		Where("good3", "=", 0).
		Where("good4", "=", 0.).
		Where("good3", "=", 2).
		Where("good5", "=", "").
		Where("good6", "like", "abc").
		Where("good7", "in", []int{1, 2, 3}).
		BuildQuery()

	log.Println(sql)
	log.Println(v)
}

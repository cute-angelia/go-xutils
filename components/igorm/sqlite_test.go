package igorm

import (
	"log"
	"testing"
)

type TestModel struct {
	Abc string `json:"abc"`
	Xyz int    `json:"xyz"`
}

// TestSqlite go test -v --rrn TestSqlite
func TestSqlite(t *testing.T) {
	New(WithDbName("cache")).MustInitSqlite()
	orm, _ := GetGormSQLite("cache")

	// Create table for `TestModel`
	orm.Migrator().CreateTable(&TestModel{})

	model := TestModel{
		Abc: "abc",
		Xyz: 11,
	}
	orm.Create(&model)

	model2 := []TestModel{}
	orm.Where("abc = ?", "abc").Find(&model2)

	log.Println(model)

	log.Println(model2)
}

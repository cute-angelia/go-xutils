package igorm

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CreateOrUpdate 创建或者更新
// ps: interface 全部传指针
func CreateOrUpdate(orm *gorm.DB, table string, data map[string]interface{}, id int32) (interface{}, error) {
	if id > 0 {
		result := orm.Table(table).Where("id = ?", id).Updates(data)
		if result.Error != nil {
			return nil, result.Error
		}
		// 不建议检查 RowsAffected == 0，因为数据无变化时也为 0
	} else {
		if err := orm.Table(table).Create(data).Error; err != nil {
			return nil, err
		}
	}
	return data, nil
}

// GetPageData 统一版
func GetPageData[T any](db *gorm.DB, page int, perPage int) ([]T, int64, error) {
	var data []T
	var total int64

	// 1. 必须先指定 Model，否则 Count 不知道查哪张表
	// 2. 使用 Session 保证隔离
	tx := db.Session(&gorm.Session{}).Model(new(T))

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * perPage
	if err := tx.Offset(offset).Limit(perPage).Find(&data).Error; err != nil {
		return nil, total, err
	}

	return data, total, nil
}

// Convert 转化数据 dest => &dest
/*
性能最优解：手动赋值（Manual Assignment）
这是性能最高、最安全的方法，完全没有反射开销，适合高频调用的接口。
// 替代 Convert
func ToUserResponse(user *UserModel) *UserResponse {
    return &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        // 这里可以自由控制哪些字段需要转换
    }
}

工业标准解：使用 copier 库
*/
func Convert(src interface{}, dest interface{}) {
	temp, _ := json.Marshal(src)
	json.Unmarshal(temp, dest)
}

// 替代你的 Convert 函数
func ConvertCopier(src interface{}, dest interface{}) error {
	return copier.Copy(dest, src)
}

// 直接在 GORM 链式调用中使用 .Select("field1", "field2") 来解决零值不更新的问题，这样连 ConvertMap 这个函数都可以省掉。
// GORM .Select()	性能最好，原生支持	需要手动写字段名	确定的模型更新
// 即使 Score 是 0，也会被强制更新
// orm.Table("users").Where("id = ?", id).Select("Score", "Name").Updates(userModel)
// ConvertMap gorm updates 对 model 为 0 的数据不处理， 这里转化为 map 对象处理
func ConvertMap(in interface{}, noKey []string) map[string]interface{} {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(in)
	json.Unmarshal(inrec, &inInterface)

	if _, ok := inInterface["id"]; ok {
		delete(inInterface, "id")
	}

	if _, ok := inInterface["uid"]; ok {
		delete(inInterface, "uid")
	}

	for _, i2 := range noKey {
		if _, ok := inInterface[i2]; ok {
			delete(inInterface, i2)
		}
	}

	for k, v := range inInterface {
		if fmt.Sprintf("%v", v) == "" {
			delete(inInterface, k)
		}
	}

	return inInterface
}

func QueryGenerate(orm *gorm.DB, key string, opt string, value interface{}) *gorm.DB {
	switch v := value.(type) {
	case string:
		if len(v) > 0 {
			if opt == "like" {
				orm = orm.Where(fmt.Sprintf("%s like ?", key), "%"+v+"%")
			} else {
				orm = orm.Where(fmt.Sprintf("%s %s ?", key, opt), v)
			}
		}
	case int:
		if v > 0 {
			orm = orm.Where(fmt.Sprintf("%s %s ?", key, opt), v)
		}
	case int32:
		if v > 0 {
			orm = orm.Where(fmt.Sprintf("%s %s ?", key, opt), v)
		}
	case []int32:
		if len(v) > 0 {
			orm = orm.Where(fmt.Sprintf("%s %s (?)", key, opt), v)
		}
	default:
		orm = orm.Where(fmt.Sprintf("%s %s ?", key, opt), v)
	}

	return orm
}

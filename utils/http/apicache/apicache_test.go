package apicache

import (
	"github.com/cute-angelia/go-xutils/components/caches/ibunt"
	"log"
	"testing"
)

func TestCache(t *testing.T) {
	u := struct {
		Id    int32
		Title string `valid:"Required;"`
		Desc  string
		Uid   int32
	}{
		Id:    1,
		Title: "good test",
		Desc:  "desc",
		Uid:   3,
	}

	ibunt.New()

	apiCache := New(ibunt.GetComponent("cache")).MustBuild("cache", "test-", WithDebug(true), WithGenerateCacheKey(u, []string{"Uid"}), WithPrefixMaxNum(5))

	log.Println("设置缓存")
	apiCache.SetCache("hello")
	apiCache.SetCache("hello2")
	apiCache.SetCache("hello3")

	log.Println("获取缓存")
	log.Println(apiCache.GetCache())

	log.Println("删除缓存")
	log.Println(apiCache.DeleteCache())

	log.Println("获取缓存")
	log.Println(apiCache.GetCache())

	log.Println("设置缓存")
	apiCache.SetCache("hello")
	log.Println("删除所有缓存")
	log.Println(apiCache.DeleteCache())
	apiCache.DeleteCacheAll()
	log.Println("获取缓存")
	log.Println(apiCache.GetCache())

}

//func TestCacheslip(t *testing.T) {
//	str := "test-abebe522541037dbab30cd577fe2c4ac|test-23a7c20e50e9a7f6792c8a2c39be7ccd|test-abebe522541037dbab30cd577fe2c4ac|test-23a7c20e50e9a7f6792c8a2c39be7ccd|test-23a7c20e50e9a7f6792c8a2c39be7ccd|test-abebe522541037dbab30cd577fe2c4ac|test-abebe522541037dbab30cd577fe2c4ac|test-71a79f9c20f2ee38c8a791f4706c7c06|test-23a7c20e50e9a7f6792c8a2c39be7ccd|test-32efdac4907b17612eda00cb816e94af|test-32efdac4907b17612eda00cb816e94af|test-32efdac4907b17612eda00cb816e94af|test-32efdac4907b17612eda00cb816e94af|test-9c111f66c3cfe3240ea9ca5d6f46e227|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0|test-cc2ff5cb494a80f9a55d4109af7d2bf0"
//
//	log.Println(str)
//
//	strs := strings.Split(str, "|")
//
//	max:=20
//	log.Println(len(strs))
//	strs = strs[len(strs)-max:len(strs)]
//	log.Println(len(strs))
//
//	log.Println(strings.Join(strs,"|"))
//}

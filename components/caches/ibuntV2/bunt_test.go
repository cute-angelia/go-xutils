package ibuntV2

import (
	"github.com/cute-angelia/go-xutils/syntax/irandom"
	"log"
	"testing"
	"time"
)

func getComponent() *Component {
	log.SetFlags(log.Lshortfile)
	return New(WithName("cache"), WithDbFile("/tmp/test.db"))
}

func TestTimeout(t *testing.T) {
	db := getComponent()
	cacheKey := "testtimeout"
	db.Set(cacheKey, irandom.RandString(10, irandom.LetterAbc), time.Second*5)
	log.Println(db.Get(cacheKey))
	log.Println(db.Get(cacheKey))
	<-time.After(time.Second * 10)
	log.Println(db.Get(cacheKey))
}

func TestBucket(t *testing.T) {
	db := getComponent()
	bucket1 := "foo"
	bucket2 := "bar"

	// 一组 2个
	key1 := db.GenerateCacheKey(bucket1, "fdsafdsafdsafd")
	key1_1 := db.GenerateCacheKey(bucket1, "fdsafdsafdsaxxxfd")

	// 一组 1个
	key2 := db.GenerateCacheKey(bucket2, "xxxfdf2")

	// 重启后。。。索引消失
	db.SetWithBucket(bucket1, key1, "foo"+irandom.RandString(10, irandom.LetterAbc), time.Minute*3)
	db.SetWithBucket(bucket1, key1_1, "foo"+irandom.RandString(10, irandom.LetterAbc), time.Minute*3)
	db.SetWithBucket(bucket2, key2, "bar"+irandom.RandString(10, irandom.LetterAbc), time.Minute*3)

	log.Println(key1)
	log.Println(db.Get(key1))
	log.Println(key1_1)
	log.Println(db.Get(key1_1))

	log.Println(key2)
	log.Println(db.Get(key2))

	err := db.Scan(bucket1, func(key string) error {
		log.Println("scan ", key)
		return nil
	})

	log.Println(err)
}

func TestLocker(t *testing.T) {
	opt1 := NewLockerOpt(WithLimit(3), WithToday(true))
	opt2 := NewLockerOpt(WithLimit(1), WithToday(false))
	opt3 := NewLockerOpt(WithLimit(35))
	for i := 1; i <= 30; i++ {
		if ok, err := IsNotLockedInLimit("cache", "test1", time.Hour, opt1); ok {
			log.Println(err)
			log.Println("opt1 > ", i)
			//2020/08/13 17:17:17 opt1 >  1
			//2020/08/13 17:17:17 opt1 >  2
			//2020/08/13 17:17:17 opt1 >  3
		}
	}

	for i := 1; i <= 30; i++ {
		if ok, err := IsNotLockedInLimit("cache", "test2", time.Hour, opt2); ok {
			log.Println(err)
			log.Println("opt2 > ", i)
			// 2020/08/13 17:17:17 opt2 >  1
		}
	}

	for i := 1; i <= 30; i++ {
		if ok, err := IsNotLockedInLimit("cache", "test3", time.Hour, opt3); ok {
			log.Println(err)
			log.Println("opt3 > ", i)
		}
	}

}

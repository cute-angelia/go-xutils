package ibitcask

import (
	"git.mills.io/prologic/bitcask"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"log"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	db, _ := bitcask.Open("/tmp/bitcask", bitcask.WithMaxDatafileSize(1024*1024*200))
	db.RunGC()
	defer db.Close()
	// db.Put([]byte("Hello"), []byte("World"))
	val, _ := db.Get([]byte("Hello"))
	log.Printf(string(val))

	db.PutWithTTL([]byte("Hello2"), []byte("World"), time.Second*2)
	val2, _ := db.Get([]byte("Hello2"))
	log.Printf(string(val2))

	<-time.After(time.Second * 2)

	val3, err := db.Get([]byte("Hello2"))
	log.Printf(string(val3))
	log.Println(err)
}

// TestComponent 测试组件形式
// go test -v --run TestComponent
func TestComponent(t *testing.T) {
	bitcaskC := New(WithPath("/tmp/bitcask"))

	log.Println(bitcaskC.Set("test", "value", time.Second*2))
	log.Println(bitcaskC.Get("test"))

	db, _ := bitcaskC.GetDb()
	stats, _ := db.Stats()
	log.Println(ijson.Pretty(stats))

	db.Scan([]byte("t"), func(key []byte) error {
		val, err := db.Get(key)
		if err != nil {
			log.Println(string(key), "error => ", err)
			return nil
		}
		log.Println(string(key), "=>", string(val))
		return nil
	})

	log.Println("-----")

	db.Fold(func(key []byte) error {
		val, err := db.Get(key)
		if err != nil {
			log.Println(string(key), "error => ", err)
			return nil
		}
		log.Println(string(key), "=>", string(val))
		return nil
	})

	<-time.After(time.Second * 2)
	log.Println(bitcaskC.Get("test"))
}

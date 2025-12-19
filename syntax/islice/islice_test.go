package islice

import (
	"fmt"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"log"
	"testing"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func TestDelete(t *testing.T) {
	var strs []string

	for i := 0; i < 10; i++ {
		strs = append(strs, fmt.Sprintf("xx_%d", i))
	}
	log.Println(ijson.Pretty(strs))
	strs = RemoveStringWithOrder(strs, 3)
	log.Println(ijson.Pretty(strs))

	strs = RemoveStringWithoutOrder(strs, 3)
	log.Println(ijson.Pretty(strs))
}

func TestInSlice(t *testing.T) {
	a := "a"
	b := "b"
	c := "c"
	d := "d"
	e := "e"
	f := "f"

	islice := NewInSlice("a", "c")

	log.Println("a:", islice.Has(a))
	log.Println("b:", islice.Has(b))
	log.Println("c:", islice.Has(c))
	log.Println("d:", islice.Has(d))
	log.Println("e:", islice.Has(e))
	log.Println("f:", islice.Has(f))

	islice = NewInSlice("d", "c")

	log.Println("a:", islice.Has(a))
	log.Println("b:", islice.Has(b))
	log.Println("c:", islice.Has(c))
	log.Println("d:", islice.Has(d))
	log.Println("e:", islice.Has(e))
	log.Println("f:", islice.Has(f))

}

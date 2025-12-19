package ierror

import (
	"errors"
	"github.com/cute-angelia/go-xutils/syntax/ijson"
	"log"
	"testing"
)

func TestError(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := New(404, "图片不存在")
	err3 := notfound()

	err4 := NewCode(IllegalUserName)
	err5 := New(20101, "图片不存在")

	log.Println("compared: errors.new vs New => ", errors.Is(err1, err2))
	log.Println("compared: New vs New=> ", errors.Is(err2, err3))
	log.Println("compared: New vs New=> ", errors.Is(err4, err5), err4.Error(), err5.Error())
	log.Println("compared: New vs nil=> ", errors.Is(err2, nil))
	log.Println("err1 => ", err1.Error())
	log.Println("err2 => ", err2.Error())
	log.Println("err3 => ", err3.Error())
	log.Println("err4 => ", err4.Error())

	// 读取code
	var restErr = new(Error)
	ok := errors.As(err4, &restErr)
	log.Println(ok, restErr.Code, restErr.Message)
	log.Println(ijson.Pretty(restErr))
	log.Printf("%d", restErr.Code)
	code := restErr.Code
	log.Println("code :", code)
	log.Println(restErr.Code, restErr.Code.String())
}

func notfound() error {
	return New(404, "图片不存在2")
}

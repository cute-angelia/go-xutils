## uid crypt

UID 转化

```
package main

import (
	"github.com/cute-angelia/go-xutils/grant/idCrypt"
	"log"
)

func main() {
	in32 := uint32(233)
	in64 := uint64(233)

	crypt := idCrypt.NewIdCrypt32()
	log.Println(crypt.Encrypt(in32))
	log.Println(crypt.Decrypt(crypt.Encrypt(in32)))

	crypt2 := idCrypt.NewIdCrypt64()
	log.Println(crypt2.Encrypt(in64))
	log.Println(crypt2.Decrypt(crypt2.Encrypt(in64)))
	
}

2018/11/28 11:43:17 1075056177
2018/11/28 11:43:17 233
2018/11/28 11:43:17 6926536228913924265
2018/11/28 11:43:17 233

```
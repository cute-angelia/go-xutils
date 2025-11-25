## apiv3


### 功能

1. 解析 request body 内容
2. 校验 request [ozzo-validation](https://github.com/go-ozzo/ozzo-validation)
3. 返回 response 内容

### 参考
ref: [render](https://github.com/go-chi/render)


### code

```go
package auth

import (
	"github.com/cute-angelia/go-utils/syntax/itime"
	"github.com/cute-angelia/go-xutils/utils/http/api"
	"github.com/cute-angelia/go-xutils/utils/http/apiV3"
	"github.com/cute-angelia/go-xutils/utils/http/ip"
	"github.com/go-chi/chi"
	"github.com/go-ozzo/ozzo-validation/v4"
	"go-admin/internal/consts"
	"go-admin/internal/consts/errorcode"
	"go-admin/internal/user"
	"go-admin/model/modelUser"
	"net/http"
	"time"
)

type Username struct {
}

func (self Username) Routes() chi.Router {
	r := chi.NewRouter()
	// 登录
	r.Post("/login", self.login)
	//r.Post("/register", self.register)
	return r
}

func (self Username) login(w http.ResponseWriter, r *http.Request) {
	resp := apiV3.NewRender(w, r)
	// 校验参数
	req := struct {
		Username string `valid:"Required;"`
		Password string `valid:"Required;"`
	}{}
	if err := resp.Decode(&req); err != nil {
		resp.Error(err)
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Username, validation.Required, validation.Length(5, 30)),
		validation.Field(&req.Password, validation.Required, validation.Length(5, 40)),
	); err != nil {
		resp.Error(err)
		return
	}

	// 公共参数
	userIp := ip.RemoteIp(r)

	accountLib := user.NewAccount(consts.BoxAppid)
	if isReged, accountiInfo := accountLib.IsRegistered(req.Username); isReged {

		// 校验密码
		if _, err := accountLib.BeforeLoginCheckPassword(req.Username, req.Password); err != nil {
			api.Error(w, r, nil, "账号或密码错误", -1)
			return
		}

		// 更新 最后登录时间 IP, MAC
		accountLib.UpdateOther(accountiInfo.Uid, map[string]interface{}{
			"last_login_time": itime.NewUnixNow().Format(),
			"last_login_ip":   userIp,
		})

		// 返回登录后信息
		resp.SetData(getUserLoginResp(accountiInfo), "登陆成功").Success()
		return
	} else {
		resp.Error(apiV3.NewApiError(int32(errorcode.ErrorLoginNotFound), errorcode.ErrorLoginNotFound.String()))
		return
	}
}
```
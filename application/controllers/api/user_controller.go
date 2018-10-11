package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-xorm/xorm"
	"github.com/imroc/req"
	jwt2 "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/segmentio/objconv/json"
	"iriscms/application/models/tables"
	"strconv"
	"time"
)

type UserApiController struct {
	Orm *xorm.Engine
	Ctx iris.Context
}

type ReturnApiData struct {
	Status bool        `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func (c *UserApiController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle(iris.MethodPost, "/user/login", "UserLogin")
	b.Handle(iris.MethodGet, "/user/center", "UserCenter")
}

func (c *UserApiController) UserLogin() {
	//生成JwtToken
	dd := map[string]string{}
	err := c.Ctx.UnmarshalBody(&dd, iris.UnmarshalerFunc(json.Unmarshal)) //todo 解析body字符串
	if err != nil {
		c.Ctx.JSON(ReturnApiData{false, err.Error(), nil})
		return
	}
	//验证vcaptchatoken是否正确
	vcaptchaData := map[string]interface{}{}
	res, err := req.Post("http://api.vaptcha.com/v2/validate",  req.Param{
		"id": "5bbc46c6fc650e3be06e5869",
		"secretkey": "1d374f7f501a4d2e9ca977dd343ffde8",
		"token": dd["token"],
	})
	if err != nil {
		c.Ctx.JSON(ReturnApiData{false, "validate captcha failed!", err.Error()})
		return
	}
	err = res.ToJSON(&vcaptchaData)
	if err != nil {
		c.Ctx.JSON(ReturnApiData{false, "validate captcha failed!", err.Error()})
		return
	}
	status, _ := vcaptchaData["success"].(int64)
	if status == 0 { //验证不通过
		c.Ctx.JSON(ReturnApiData{false, "validate captcha failed,status: " + strconv.Itoa(int(status)) + " msg:" + vcaptchaData["msg"].(string) + "!", nil})
		return
	}
	if dd["account"] == "" || dd["password"] == "" {
		c.Ctx.JSON(ReturnApiData{false, "username or password is empty!", nil})
	} else {
		var user tables.IriscmsMember
		ok, err := c.Orm.Where("account = ? and password = ?", dd["account"], dd["password"]).Get(&user)
		if !ok || err != nil {
			c.Ctx.JSON(ReturnApiData{false, "login failed!", nil})
			return
		}

		claims := jwt.MapClaims{ //为了匹配中间件, 在这里使用相同的配置
			"user": user,
			"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(), //过期时间
			"iat":  time.Now().Unix(),                                   // 当前时间戳
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("MySecret"))
		c.Ctx.JSON(ReturnApiData{true, "login success!", struct {
			SignToken string `json:"sign_token"`
		}{SignToken: tokenString}})
	}
}

func (c *UserApiController) UserCenter() {
	user, ok := c.Ctx.Values().Get(jwt2.DefaultContextKey).(*jwt.Token)
	if !ok {
		c.Ctx.JSON(ReturnApiData{false, "author error", nil})
		return
	}
	userC, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		c.Ctx.JSON(ReturnApiData{false, "author error", nil})
		return
	}
	c.Ctx.JSON(ReturnApiData{true, "", userC["user"]})
}

func (c *UserApiController) UserLogout() {
	c.Ctx.JSON(ReturnApiData{true, "", nil})
}
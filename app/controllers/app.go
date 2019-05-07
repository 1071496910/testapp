package controllers

import (
	"encoding/json"
	"github.com/revel/revel"
	"github.com/revel/revel/logger"
	"hash/adler32"
	"net/http"
	"net/url"
	"testapp/app"
	"time"
)

const (
	registerSQL      = `INSERT INTO account ( Name, email, password, create_time) VALUES (?, ?, ?, ?);`
	loginSQL         = `SELECT count(Name) FROM account WHERE Name = ? and password = ? ;`
	cookieLoginKey   = "lastLoginInfo"
	customCookieName = "REVEL_CUSTOM_COOKIES"
)

type App struct {
	*revel.Controller
}

type loginInfo struct {
	Name  string    `json:"name"`
	Token uint32    `json:"token"`
	Time  time.Time `json:"time"`
}

func checkLogin(logger logger.MultiLogger, c *revel.Controller) bool {

	cookie, err := c.Request.Cookie(customCookieName)
	if err != nil {
		return false
	}
	loginInfoStr, err := url.QueryUnescape(cookie.GetValue())
	//loginInfoStr := cookie.GetValue()
	logger.Debugf("In checkLogin() get cookie %v \n.", loginInfoStr)
	if err != nil {
		return false
	}

	if loginInfoStr == "" {
		logger.Warnf("cookie is empty\n")
		return false
	}
	lastLoginInfo := &loginInfo{}
	if err := json.Unmarshal([]byte(loginInfoStr), lastLoginInfo); err != nil {
		logger.Warnf("parse cookie error: %v\n", err)
		return false
	}

	lastLoginTime := lastLoginInfo.Time
	expireTime := lastLoginTime.Add(time.Hour)

	if lastLoginTime.After(expireTime) {
		return false
	}

	if lastLoginInfo.Token != app.TokenMap[lastLoginInfo.Name] {
		logger.Warnf("User: %v token[%v] is not math server token[%v].\n",
			lastLoginInfo.Name,
			lastLoginInfo.Token,
			app.TokenMap[lastLoginInfo.Name])
		return false
	}

	lastLoginInfo.Time = time.Now()
	data, err := json.Marshal(lastLoginInfo)
	if err != nil {
		return false
	}
	c.SetCookie(&http.Cookie{
		HttpOnly: false,
		Name:     customCookieName,
		Value:    url.QueryEscape(string(data)),
		Path:     "/",
	})

	return true
}

func (c App) Index() revel.Result {

	if checkLogin(c.Log, c.Controller) {
		return c.Redirect("/blog/editor")
	}

	return c.Render()
}

func (c App) Register() revel.Result {
	return c.Render()
}

func (c App) DoLogin(name, password string) revel.Result {
	delete(c.Flash.Out, "error")

	var matched int64
	var data []byte
	var err error
	var token uint32
	var loginTime time.Time

	stmt, err := app.DB.Prepare(loginSQL)
	if err != nil {
		goto ErrHandler
	}
	defer stmt.Close()

	err = stmt.QueryRow(name, password).Scan(&matched)
	if err != nil {
		goto ErrHandler
	}

	if matched == 0 {
		c.Flash.Error("Account not exists or password error!")
		c.Log.Infof("User : %v login error, password is  %v .\n", name, password)
		return c.Redirect("/")
	}

	loginTime = time.Now()
	token = adler32.Checksum([]byte(name + loginTime.String()))

	data, err = json.Marshal(&loginInfo{
		Name:  name,
		Time:  loginTime,
		Token: token,
	})
	if err != nil {
		return c.RenderError(err)
	}
	c.Log.Debugf("Get json: %v \n.", string(data))
	c.SetCookie(&http.Cookie{
		HttpOnly: false,
		Name:     "REVEL_CUSTOM_COOKIES",
		Value:    url.QueryEscape(string(data)),
		Path:     "/",
	})

	c.Log.Debugf("Set user: %v token: %v.\n", name, token)
	app.TokenMap[name] = token

	return c.Redirect("/blog/editor")

ErrHandler:
	return c.RenderError(err)
}

func (c App) GetSession() revel.Result {
	val, err := c.Session.Get("test")
	if err != nil {
		return c.RenderJSON(err)
	}
	return c.RenderJSON(val)
	//return c.RenderJSON(c.Session["test"])
}

func (c App) SetSession(sv string) revel.Result {
	c.Session.Set("test", sv)
	return c.RenderJSON("")
}

func (c App) Test() revel.Result {
	return c.Render()
}

func (c App) DoReg(name, email, password string) revel.Result {
	/*stmt, err := app.DB.Prepare(registerSQL)
	if err != nil {
		goto ErrHandler
	}
	defer stmt.Close()

	if _, err = stmt.Exec(Name, email, password, Time.Now()); err != nil {
		goto ErrHandler
	}*/

	return c.Render(name, password)

	/*	ErrHandler:
		return c.RenderError(err)*/
}

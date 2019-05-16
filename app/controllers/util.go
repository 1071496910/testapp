package controllers

import (
	"encoding/json"
	"github.com/revel/revel"
	"github.com/revel/revel/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testapp/app"
	"time"
)

func init() {

	revel.InterceptFunc(CheckLogin, revel.BEFORE, Blog{})
	revel.InterceptFunc(CheckLogin, revel.BEFORE, Image{})
}

// simple example or user auth
func CheckLogin(c *revel.Controller) revel.Result {
	if !checkLogin(c.Log, c) {
		return c.Redirect("/")
	}
	return nil
}

func checkLogin(logger logger.MultiLogger, c *revel.Controller) bool {

	lastLoginInfo, err := GetLoginInfo(logger, c)
	if err != nil {
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
		Name:     app.CustomCookieName,
		Value:    url.QueryEscape(string(data)),
		Path:     "/",
	})

	return true
}

type ArtInfo struct {
	ID       string
	Owner    string
	Title    string
	Location string
	Content  string
}

const findByIdSQL = `select  owner, title, location from article where id = ? ;`

func GetArtInfoByID(id string) (*ArtInfo, error) {
	if id == "" {
		return nil, ErrIDEmpty
	}
	stmt, err := app.DB.Prepare(findByIdSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	artInfo := &ArtInfo{}
	artInfo.ID = id

	if rows.Next() {
		err = rows.Scan(&artInfo.Owner, &artInfo.Title, &artInfo.Location)
		if err != nil {
			return nil, err
		}
	}

	fp, err := os.Open(filepath.Join(app.ArticlePathPrefix, artInfo.Owner, artInfo.Location))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	artInfo.Content = string(data)

	return artInfo, nil
}

func GetLoginInfo(logger logger.MultiLogger, c *revel.Controller) (*LoginInfo, error) {
	cookie, err := c.Request.Cookie(app.CustomCookieName)
	if err != nil {
		return nil, err
	}
	LoginInfoStr, err := url.QueryUnescape(cookie.GetValue())
	//LoginInfoStr := cookie.GetValue()
	logger.Debugf("In CheckLogin() get cookie %v \n.", LoginInfoStr)
	if err != nil {
		return nil, err
	}

	if LoginInfoStr == "" {
		logger.Warnf("cookie is empty\n")
		return nil, err
	}
	lastLoginInfo := &LoginInfo{}
	if err := json.Unmarshal([]byte(LoginInfoStr), lastLoginInfo); err != nil {
		logger.Warnf("parse cookie error: %v\n", err)
		return nil, err
	}

	return lastLoginInfo, nil
}

type LoginInfo struct {
	Name  string    `json:"name"`
	Token uint32    `json:"token"`
	Time  time.Time `json:"time"`
}

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

const (
	cookieLoginKey = "lastLoginInfo"
)

type artInfo struct {
	ID       string
	Owner    string
	Title    string
	Location string
	Content  string
}

func GetArtInfoByID(id string) (*artInfo, error) {
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

	artInfo := &artInfo{}
	artInfo.ID = id

	if rows.Next() {
		err = rows.Scan(&artInfo.Owner, &artInfo.Title, &artInfo.Location)
		if err != nil {
			return nil, err
		}
	}

	fp, err := os.Open(filepath.Join(articlePathPrefix, artInfo.Owner, artInfo.Location))
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

func getLoginInfo(logger logger.MultiLogger, c *revel.Controller) (*loginInfo, error) {
	cookie, err := c.Request.Cookie(customCookieName)
	if err != nil {
		return nil, err
	}
	loginInfoStr, err := url.QueryUnescape(cookie.GetValue())
	//loginInfoStr := cookie.GetValue()
	logger.Debugf("In checkLogin() get cookie %v \n.", loginInfoStr)
	if err != nil {
		return nil, err
	}

	if loginInfoStr == "" {
		logger.Warnf("cookie is empty\n")
		return nil, err
	}
	lastLoginInfo := &loginInfo{}
	if err := json.Unmarshal([]byte(loginInfoStr), lastLoginInfo); err != nil {
		logger.Warnf("parse cookie error: %v\n", err)
		return nil, err
	}

	return lastLoginInfo, nil
}

func checkLogin(logger logger.MultiLogger, c *revel.Controller) bool {

	lastLoginInfo, err := getLoginInfo(logger, c)
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
		Name:     customCookieName,
		Value:    url.QueryEscape(string(data)),
		Path:     "/",
	})

	return true
}

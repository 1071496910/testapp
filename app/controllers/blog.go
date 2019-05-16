package controllers

import (
	"errors"
	"github.com/revel/revel"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testapp/app"
)

const (
	findByTagSQL = `select article.title
from article right join art_tag on article.id = art_tag.art_id and  article.owner = ? and art_tag.art_tag = ? where article.title is not NULL;`

	addNewFileSQL = `insert into article (owner, title, location) values ('?', '?', 'nil');`
	updateFileSQL = `update article set title=? , location=?  where id=?;`
)

var (
	//articlePathPrefix = "//dav.jianguoyun.com@SSL/DavWWWRoot/dav/blog"

	ErrIDEmpty = errors.New("article id is empty")
)

type Blog struct {
	*revel.Controller
}

func (c Blog) Home(name string) revel.Result {

	return c.Render()
}

func (c Blog) Editor(article string) revel.Result {

	artInfo, err := GetArtInfoByID(article)
	if err != nil {
		return c.RenderError(err)
	}

	c.Log.Debugf("Get artInfo %v\n", artInfo)

	c.ViewArgs["title"] = artInfo.Title
	return c.Render(article)
}

func (c Blog) ArticleByTag(owner string, tag string) revel.Result {
	stmt, err := app.DB.Prepare(findByTagSQL)
	if err != nil {
		return c.RenderError(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(owner, tag)
	if err != nil {
		return c.RenderError(err)
	}
	result := []string{}
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		c.Log.Debugf("In loop , get title %v\n", tmp)
		if err != nil {
			return c.RenderError(err)
		}
		result = append(result, tmp)
	}
	c.Log.Debugf("Get result %v \n", result)

	c.ViewArgs["result"] = result

	return c.Render(result)
}

func (c Blog) Article(article string) revel.Result {
	if article == "" {
		return c.RenderJSON("")
	}

	artInfo, err := GetArtInfoByID(article)
	if err != nil {
		return c.RenderError(err)
	}
	return c.RenderJSON(artInfo.Content)
}

func (c Blog) Save(title, id, md string) revel.Result {

	c.Log.Debugf("title: %v, id %v, md %v.\n", title, id, md)

	lastLoginInfo, err := GetLoginInfo(c.Log, c.Controller)
	if err != nil {
		return c.RenderError(err)
	}
	c.Log.Debug(c.Request.Form.Encode())

	var validID int64

	if err != nil {

		if err == ErrIDEmpty {

			stmt, err := app.DB.Prepare(addNewFileSQL)
			if err != nil {
				return c.RenderError(err)
			}
			defer stmt.Close()

			rest, err := stmt.Exec(lastLoginInfo.Name, title)
			if err != nil {
				return c.RenderError(err)
			}
			validID, err = rest.LastInsertId()
			if err != nil {
				return c.RenderError(err)
			}
		} else {
			return c.RenderError(err)
		}
	} else {
		validID, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return c.RenderError(err)
		}
	}

	id = strconv.FormatInt(validID, 10)

	dir := filepath.Join(app.ArticlePathPrefix, lastLoginInfo.Name)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return c.RenderError(err)
	}
	path := filepath.Join(dir, id)
	c.Log.Debugf("Wirte to file %v .\n", path)
	ioutil.WriteFile(path, []byte(md), 0644)

	updateStmt, err := app.DB.Prepare(updateFileSQL)
	if err != nil {
		return c.RenderError(err)
	}
	defer updateStmt.Close()
	result, err := updateStmt.Exec(title, id, id)
	if err != nil {
		return c.RenderError(err)
	}
	n, err := result.RowsAffected()

	if err != nil {
		return c.RenderError(err)
	}
	c.Log.Debugf("Affected %v rows.\n", n)
	return c.Redirect("/blog/home")
}

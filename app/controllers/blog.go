package controllers

import (
	"database/sql"
	"errors"
	"github.com/revel/revel"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testapp/app"
)

const (
	findByTagSQL = `select article.id,article.title
from article right join art_tag on article.id = art_tag.art_id and  article.owner = ? and art_tag.art_tag = ? where article.title is not NULL;`
	ListSQL = `select id,title from article where owner=? ;`

	addNewFileSQL = `insert into article (owner, title, location) values (?, ?, 'nil');`
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
	loginInfo,err := GetLoginInfo(c.Log,c.Controller)
	if err != nil {
	    return c.RenderError(err)
	}
	c.ViewArgs["owner"] = loginInfo.Name
	return c.Render()
}

func (c Blog) Editor(article string) revel.Result {

	artInfo, err := GetArtInfoByID(article)
	if err != nil {
		if err == ErrIDEmpty {
			return c.Render()
		}

		return c.RenderError(err)
	}

	c.Log.Debugf("Get artInfo %v\n", artInfo)
	c.ViewArgs["title"] = artInfo.Title
	return c.Render(article)
}

func (c Blog) ArticleByTag(tag string) revel.Result {

	loginInfo, err := GetLoginInfo(c.Log, c.Controller)
	if err != nil {
	    return c.RenderError(err)
	}
	owner := loginInfo.Name

	stmt := &sql.Stmt{}
	if tag != "" {
		stmt, err = app.DB.Prepare(findByTagSQL)
	} else {
		stmt, err = app.DB.Prepare(ListSQL)
	}
	if err != nil {
		return c.RenderError(err)
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	if tag != "" {
		rows, err = stmt.Query(owner, tag)
	}  else {
		rows, err = stmt.Query(owner)
	}
	if err != nil {
		return c.RenderError(err)
	}
	result := map[string]string{}
	for rows.Next() {
		id := ""
		title := ""
		err = rows.Scan(&id, &title)
		c.Log.Debugf("In loop , get id %v, get title %v\n", id, title)
		if err != nil {
			return c.RenderError(err)
		}
		result[id] = title
	}
	c.Log.Debugf("Get result %v \n", result)

	c.ViewArgs["result"] = result

	return c.Render()
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

func (c Blog) Read(id string) revel.Result {
	artInfo, err := GetArtInfoByID(id)

	if err != nil {
	    return c.RenderError(err)
	}
	c.ViewArgs["artHTML"] = artInfo.HTML
	c.ViewArgs["owner"] = artInfo.Owner
	return c.Render()
}

func (c Blog) Save(title, id, md, html string) revel.Result {

	c.Log.Debugf("title: %v, id %v, md %v.\n", title, id, md, html)

	lastLoginInfo, err := GetLoginInfo(c.Log, c.Controller)
	if err != nil {
		return c.RenderError(err)
	}
	c.Log.Debug(c.Request.Form.Encode())

	var validID int64

	_, err = GetArtInfoByID(id)
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
	err = ioutil.WriteFile(path, []byte(md), 0644)
	if err != nil {
	    return c.RenderError(err)
	}
	err = ioutil.WriteFile(path + ".html", []byte(html), 0644)
	if err != nil {
		return c.RenderError(err)
	}

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
	return nil
}

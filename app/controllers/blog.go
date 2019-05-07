package controllers

import (
	"github.com/revel/revel"
	"testapp/app"
)

const (
    findByTagSQL = `select article.title
from article right join art_tag on article.id = art_tag.art_id and art_tag.art_tag = ? and article.owner = ?;`
)

type Blog struct {
	*revel.Controller
}

func (c Blog) Home(name string) revel.Result {
	if !checkLogin(c.Log, c.Controller) {
		return c.Redirect("/")
	}
	return c.Render()
}

func (c Blog) Editor() revel.Result {
	if !checkLogin(c.Log, c.Controller) {
		return c.Redirect("/")
	}
	return c.Render()
}

func (c Blog) ArticleByTag(owner string, tag string) revel.Result {
	stmt, err := app.DB.Prepare(findByTagSQL)
	if err != nil {
		return c.Render("")
	}
	rows, err := stmt.Query(owner, tag)
	if err != nil {
		return c.Render("")
	}
	result := []string
	err = rows.Scan(result)
	if err != nil {
		return c.Render("")
	}
	c.RenderText("")
}

func (c Blog) Article(id string) revel.Result {
	if id == "" {
		return c.RenderJSON("")
	}
	return c.RenderJSON("")
}

func (c Blog) Save() revel.Result {
	c.Log.Debug(c.Request.Form.Encode())
	return c.RenderJSON("")
}

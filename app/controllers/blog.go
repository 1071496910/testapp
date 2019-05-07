package controllers

import (
	"github.com/revel/revel"
	"testapp/app"
)

const (
    findByTagSQL = `select article.title
from article right join art_tag on article.id = art_tag.art_id and  article.owner = ? and art_tag.art_tag = ? where article.title is not NULL;`
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
		return c.RenderError(err)
	}
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

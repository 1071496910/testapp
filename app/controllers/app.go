package controllers

import (
	"github.com/revel/revel"
	"testapp/app"
)

const (
	registerSQL = `INSERT INTO account ( name, email, password, create_time) VALUES (?, ?, ?, ?);`
	loginSQL = `SELECT count(name) FROM account WHERE name = ? and password = ? ;`
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	c.FlashParams()
	return c.Render()
}

func (c App)Register() revel.Result  {
	return c.Render()
}

func (c App)DoLogin(name, password string) revel.Result {
	delete(c.Flash.Out, "error")

	var matched int64

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
		c.Log.Info("User :" + name, "login error, password is " +  password + ".")
		return c.Redirect("/")
	}

	return c.Render(name,password)

ErrHandler:
	return c.RenderError(err)
}

func (c App)DoReg(name, email, password string) revel.Result  {
	/*stmt, err := app.DB.Prepare(registerSQL)
	if err != nil {
		goto ErrHandler
	}
	defer stmt.Close()

	if _, err = stmt.Exec(name, email, password, time.Now()); err != nil {
		goto ErrHandler
	}*/

	return c.Render(name,password)

/*	ErrHandler:
		return c.RenderError(err)*/
}
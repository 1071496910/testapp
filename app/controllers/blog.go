package controllers

import (
	"github.com/revel/revel"
)

const (

)

type Blog struct {
	*revel.Controller
}

func (B Blog) Home(name string) revel.Result {
	return B.Render()
}
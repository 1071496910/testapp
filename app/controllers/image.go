package controllers

import (
	"github.com/revel/revel"
	"os"
	"path/filepath"
)

type Image struct {
	*revel.Controller
}

func (c Image) Upload(artID string, imageName string, data []byte) revel.Result {

	imgDir := filepath.Join("/public/", artID)
	err := os.MkdirAll(imgDir, 0755)
	if err != nil {
		c.Log.Error("Mkdir() error: %v\n", err)
		return c.RenderError(err)
	}

	file := filepath.Join(imgDir, imageName)
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		c.Log.Error("OpenFile() error: %v\n", err)
		return c.RenderError(err)
	}
	defer fp.Close()

	_, err = fp.Write(data)
	if err != nil {
		c.Log.Error("Write() error: %v\n", err)
		return c.RenderError(err)
	}
	return c.RenderText(file)
}

func (c Image) Delete(file string) revel.Result {

	err := os.Remove(file)
	if err != nil {
		c.RenderError(err)
	}
	return c.RenderText("")
}

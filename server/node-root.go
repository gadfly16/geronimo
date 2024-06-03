package server

import "github.com/gin-gonic/gin"

type Root struct {
	DetailModel
}

func (root *Root) displayData() (display gin.H) {
	display = gin.H{}
	return
}

func (root *Root) run() error {
	return nil
}

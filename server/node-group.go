package server

import "github.com/gin-gonic/gin"

type Group struct {
	DetailModel
}

func (group *Group) displayData() (display gin.H) {
	display = gin.H{}
	return
}

func (group *Group) run() error {
	return nil
}

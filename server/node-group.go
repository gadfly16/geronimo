package server

import "github.com/gin-gonic/gin"

type Group struct {
	DetailModel
}

func (group *Group) DisplayData() (display gin.H) {
	display = gin.H{}
	return
}

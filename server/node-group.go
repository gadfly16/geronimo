package server

import "github.com/gin-gonic/gin"

type Group struct {
	DetailModel
}

func (group *Group) Display() (detail gin.H) {
	detail = gin.H{}
	detail["Type"] = "group"
	return
}

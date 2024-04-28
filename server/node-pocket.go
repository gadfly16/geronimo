package server

import "github.com/gin-gonic/gin"

type Pocket struct {
	DetailModel
	Holdings map[Asset]float64

	account *Account
}

func (pocket *Pocket) Display() (detail gin.H) {
	detail = gin.H{}
	detail["Type"] = "pocket"
	detail["Holdings"] = pocket.Holdings
	return
}

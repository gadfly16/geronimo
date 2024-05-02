package server

import "github.com/gin-gonic/gin"

type Pocket struct {
	DetailModel
	Holdings map[Asset]float64

	account *Account
}

func (pocket *Pocket) DisplayData() (display gin.H) {
	display = gin.H{}
	display["Holdings"] = pocket.Holdings
	return
}

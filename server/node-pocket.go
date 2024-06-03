package server

import "github.com/gin-gonic/gin"

type Pocket struct {
	DetailModel
	Holdings map[Asset]float64

	account *Account
}

func (pocket *Pocket) displayData() (display gin.H) {
	display = gin.H{}
	display["Holdings"] = pocket.Holdings
	return
}

func (pocket *Pocket) run() error {
	return nil
}

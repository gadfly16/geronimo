package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type APIError struct {
	Error string
}

func (core *Core) apiRoutes(r *gin.Engine) {
	api := r.Group("/api", core.needUserRole)
	{
		api.GET(APIAccount, core.getAccount)
		api.POST(APIAccount, core.postAccount)
		api.GET(APIState, core.getState)
	}
}

func (core *Core) getAccount(c *gin.Context) {

}

func (core *Core) postAccount(c *gin.Context) {
	accNode := &Node{
		Detail: &Account{},
	}
	var err error

	if err = c.ShouldBindJSON(accNode); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	userRole, _ := c.Get("role")
	userID, _ := c.Get("userID")
	if userRole != "admin" && userID != strconv.Itoa(int(accNode.ParentID)) {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	msg := &Message{
		Type:     MessageCreateAccount,
		Payload:  accNode,
		RespChan: make(chan *Message),
	}

	core.message <- msg
	resp := <-msg.RespChan
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	log.Debugf("Account created: %+v\n", accNode)
}

func (core *Core) getState(c *gin.Context) {
	userRole, _ := c.Get("role")
	userID, _ := c.Get("userID")
	reqUserID := c.Query("userid")
	if userRole != "admin" && userID != reqUserID {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	uid, err := strconv.Atoi(reqUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	resp := core.send(MessageGetState, uint(uid))
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	c.String(http.StatusOK, string(resp.JSPayload))
}

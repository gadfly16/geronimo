package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	APITree    = "/tree"
	APIDetail  = "/detail"
	APIAccount = "/account"
	APICreate  = "/create"
)

type APIError struct {
	Error string
}

func (core *Core) apiRoutes(r *gin.Engine) {
	api := r.Group("/api", needUserRole)
	{
		api.GET(APITree, getTree)
		api.GET(APIDetail+"/*path", getDetail)
		api.POST(APICreate+"/:objtype", createAPIHandler)
		api.POST(APIAccount, postAccount)
	}
}

func getRequestUser(c *gin.Context) (user *User) {
	return core.nodes[c.GetUint("userID")].Detail.(*User)
}

func createAPIHandler(c *gin.Context) {
	// body, _ := io.ReadAll(c.Request.Body)
	// log.Debug(string(body))
	user := getRequestUser(c)
	objType := c.Param("objtype")
	node := &Node{}
	switch objType {
	case "broker":
		node.Detail = &Broker{}
	}
	msg := &Message{Payload: node}
	if err := c.BindJSON(msg); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	msg.User = user
	resp := msg.toCore()
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	log.Debugf("%+v %+v %+v %+v", objType, msg, msg.Payload, msg.Payload.(*Node).Detail)
}

func getDetail(c *gin.Context) {
	reqUser := getRequestUser(c)
	path := c.Param("path")
	node := find(core.root, path, reqUser)
	if node == nil {
		c.JSON(http.StatusBadRequest, APIError{"can't find node"})
		return
	}
	log.Printf("Path param: %s", path)
	c.JSON(http.StatusOK, node)
}

func postAccount(c *gin.Context) {
	user := getRequestUser(c)
	accNode := &Node{
		Detail: &Account{},
	}
	var err error

	if err = c.ShouldBindJSON(accNode); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	if user.Role != "admin" && user.NodeID != accNode.ParentID {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	resp := core.send(MessageCreateAccount, accNode)
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	log.Debugf("Account created: %+v\n", accNode)
}

func getTree(c *gin.Context) {
	user := getRequestUser(c)

	queryUserID, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	if user.Role != "admin" && int(user.NodeID) != queryUserID {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	resp := core.send(MessageGetState, uint(queryUserID))
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	c.String(http.StatusOK, string(resp.Payload.([]byte)))
}

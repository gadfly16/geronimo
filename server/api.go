package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	APITree    = "/tree"
	APIDisplay = "/display"
	APIAccount = "/account"
	APICreate  = "/create"
	APIUpdate  = "/update"
)

type APIError struct {
	Error string
}

func (core *Core) apiRoutes(r *gin.Engine) {
	api := r.Group("/api", needUserRole)
	{
		api.GET(APITree, getTree)
		api.GET(APIDisplay+"/:id", getDisplayAPIHandler)
		api.POST(APICreate+"/:objtype", createAPIHandler)
		api.POST(APIUpdate+"/:objtype", updateAPIHandler)
	}
}

func getRequestUser(c *gin.Context) (user *Node) {
	return core.nodes[c.GetUint("userID")]
}

func updateAPIHandler(c *gin.Context) {
	user := getRequestUser(c)
	objType := c.Param("objtype")
	msg := &Message{}
	switch objType {
	case "broker":
		msg.Payload = &Broker{}
	case "account":
		msg.Payload = &Account{}
	case "group":
		msg.Payload = &Group{}
	case "pocket":
		msg.Payload = &Pocket{}
	default:
		c.JSON(http.StatusBadRequest, APIError{Error: "unknown object type: " + objType})
		return
	}
	if err := c.BindJSON(msg); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	msg.User = user
	msg.Type = MessageUpdate
	resp := msg.toCore()
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	// log.Debugf("%+v %+v %+v %+v", objType, msg, msg.Payload, msg.Payload.(*Node).Detail)
}

func createAPIHandler(c *gin.Context) {
	user := getRequestUser(c)
	objType := c.Param("objtype")
	node := &Node{}
	switch objType {
	case "broker":
		node.Detail = &Broker{}
	case "account":
		node.Detail = &Account{}
	case "group":
		node.Detail = &Group{}
	case "pocket":
		node.Detail = &Pocket{}
	default:
		c.JSON(http.StatusBadRequest, APIError{Error: "unknown object type: " + objType})
		return
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

func getDisplayAPIHandler(c *gin.Context) {
	msg := &Message{Type: MessageGetDisplay}
	msg.User = getRequestUser(c)
	msg.Payload = c.Param("id")
	resp := msg.toCore()
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	c.JSON(http.StatusOK, resp.Payload)
}

func getTree(c *gin.Context) {
	user := getRequestUser(c)
	userDetail := user.Detail.(*User)

	queryUserID, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	if userDetail.Role != "admin" && int(userDetail.NodeID) != queryUserID {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	resp := core.send(MessageGetTree, uint(queryUserID))
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	c.String(http.StatusOK, string(resp.Payload.([]byte)))
}

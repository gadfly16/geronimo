package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	APIAccount = "/account"
	APITree    = "/tree"
	APIDetail  = "/detail"
)

type APIError struct {
	Error string
}

type requestUser struct {
	userID uint
	role   string
}

func (core *Core) apiRoutes(r *gin.Engine) {
	api := r.Group("/api", core.needUserRole)
	{
		api.POST(APIAccount, core.postAccount)
		api.GET(APITree, core.getTree)
		api.GET(APIDetail+"/*path", core.getDetail)
	}
}

func getRequestUser(c *gin.Context) (reqUser requestUser) {
	reqUser.role = c.GetString("role")
	reqUser.userID = c.GetUint("userID")
	return
}

func (core *Core) findNode(parent *Node, path []string, user requestUser) (node *Node, err error) {
	node, ok := parent.children[path[0]]
	if !ok {
		return
	}
	if node.DetailType == NodeUser {
		if user.role != "admin" && node.ID != user.userID {
			return nil, errors.New("not authorized")
		}
	}
	if len(path) > 1 {
		return core.findNode(node, path[1:], user)
	}
	return
}

func (core *Core) getDetail(c *gin.Context) {
	reqUser := getRequestUser(c)
	path := strings.Split(c.Param("path"), "/")[1:]
	detail, err := core.findNode(core.root, path, reqUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusBadRequest, APIError{"can't find node"})
		return
	}
	log.Printf("Path param: %s", path)
	c.JSON(http.StatusOK, detail)
}

func (core *Core) postAccount(c *gin.Context) {
	reqUser := getRequestUser(c)
	accNode := &Node{
		Detail: &Account{},
	}
	var err error

	if err = c.ShouldBindJSON(accNode); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	if reqUser.role != "admin" && reqUser.userID != accNode.ParentID {
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

func (core *Core) getTree(c *gin.Context) {
	reqUser := getRequestUser(c)

	queryUserID, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	if reqUser.role != "admin" && int(reqUser.userID) != queryUserID {
		c.JSON(http.StatusMethodNotAllowed, APIError{Error: "method not allowed"})
		return
	}

	resp := core.send(MessageGetState, uint(queryUserID))
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	c.String(http.StatusOK, string(resp.JSPayload))
}

package server

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type messageHandler func(*Core, *Message) *Message

var messageHandlers = map[string]messageHandler{
	MessageCreateAccount:    createAccountHandler,
	MessageGetState:         getStateHandler,
	MessageAuthenticateUser: authenticateUserHandler,
	MessageCreateUser:       createUserHandler,
}

func createUserHandler(core *Core, msg *Message) (resp *Message) {
	node := msg.Payload.(*Node)
	user := node.Detail.(*User)
	var userExists bool
	err := core.db.Model(User{}).Select("count(*)>0").Where("email = ?", user.Email).First(&userExists).Error
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	if userExists {
		return errorMessage(http.StatusBadRequest, "user already exists")
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, "could not generate password hash")
	}
	user.Role = RoleUser

	// Add user to database and core
	err = core.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(node).Error; err != nil {
			return err
		}
		user.NodeID = node.ID
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	user.Password = ""
	node.children = make(map[string]*Node)
	node.parent = core.root
	node.Detail = user
	core.nodes[node.ID] = node
	core.root.children[node.Name] = node

	return &Message{Type: MessageOK}
}

func authenticateUserHandler(core *Core, msg *Message) (resp *Message) {
	user := msg.Payload.(*User)
	dbUser := &User{}
	if err := core.db.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	if dbUser.ID == 0 {
		return errorMessage(http.StatusBadRequest, "user does not exist")
	}

	if err := compareHashPassword(user.Password, dbUser.Password); err != nil {
		return errorMessage(http.StatusBadRequest, err.Error())
	}

	return &Message{Type: MessageUser, Payload: dbUser}
}

func createAccountHandler(core *Core, msg *Message) (resp *Message) {
	var err error
	node := msg.Payload.(*Node)
	acc := node.Detail.(*Account)

	var accExists bool
	err = core.db.Model(node).
		Select("count(*)>0").
		Where("detail_type = ? AND parent_id = ? AND name = ?", node.DetailType, node.ParentID, node.Name).
		First(&accExists).Error
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	if accExists {
		return errorMessage(http.StatusBadRequest, "account already exists")
	}

	acc.ApiPublicKey, err = encryptString(core.dbKey, node.Name, acc.ApiPublicKey)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	acc.ApiPrivateKey, err = encryptString(core.dbKey, node.Name, acc.ApiPrivateKey)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	// Add account to database and core
	err = core.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(node).Error; err != nil {
			return err
		}
		acc.NodeID = node.ID
		if err := tx.Create(acc).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	node.children = make(map[string]*Node)
	node.parent = core.nodes[node.ParentID]
	node.parent.children[node.Name] = node
	node.Detail = acc
	core.nodes[node.ID] = node

	return &Message{Type: MessageOK}
}

func getStateHandler(c *Core, msg *Message) (resp *Message) {
	var err error
	resp = &Message{Type: MessageState}

	resp.JSPayload, err = json.Marshal(c.nodes[msg.Payload.(uint)].treeMap())
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	return
}

package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type messageHandler func(*Message) *Message

var messageHandlers = map[string]messageHandler{
	MessageGetTree:          getTreeHandler,
	MessageAuthenticateUser: authenticateUserHandler,
	MessageCreateUser:       createUserHandler,
	MessageCreate:           createHandler,
	MessageUpdate:           updateHandler,
	MessageGetDisplay:       getDisplayHandler,
}

func getDisplayHandler(msg *Message) (resp *Message) {
	resp = &Message{Type: MessageDisplay}
	id, err := strconv.Atoi(msg.Payload.(string))
	if err != nil {
		return errorMessage(http.StatusBadRequest, "invalid node id in selection")
	}
	n, err := core.getNode(uint(id), msg.User)
	if err != nil {
		return errorMessage(http.StatusBadRequest, err.Error())
	}
	resp.Payload = n.display()
	return
}

func createUserHandler(msg *Message) (resp *Message) {
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

func authenticateUserHandler(msg *Message) (resp *Message) {
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

func createHandler(msg *Message) (resp *Message) {
	var err error
	node := msg.Payload.(*Node)

	if msg.Path == "" {
		return errorMessage(http.StatusBadRequest, "can't create node without a path")
	}
	if find(core.root, msg.Path, msg.User) != nil {
		return errorMessage(http.StatusBadRequest, "node already exists")
	}
	parent := findParent(core.root, msg.Path, msg.User)
	if parent == nil {
		return errorMessage(http.StatusBadRequest, "parent doesn't exists")
	}
	node.ParentID = parent.ID
	node.Name = name(msg.Path)

	err = core.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(node).Error; err != nil {
			return err
		}
		switch obj := node.Detail.(type) {
		case *Broker:
			obj.NodeID = node.ID

			acc := parent.findUpstreamClass(NodeAccount)
			if acc == nil {
				return errors.New("node must have an account upstream")
			}
			obj.account = acc.Detail.(*Account)

			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Account:
			obj.APIPublicKey, err = encryptString(core.dbKey, node.Name, obj.APIPublicKey)
			if err != nil {
				return err
			}
			obj.APIPrivateKey, err = encryptString(core.dbKey, node.Name, obj.APIPrivateKey)
			if err != nil {
				return err
			}

			obj.NodeID = node.ID
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Group:
			obj.NodeID = node.ID
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Pocket:
			obj.NodeID = node.ID

			acc := parent.findUpstreamClass(NodeAccount)
			if acc == nil {
				return errors.New("node must have an account upstream")
			}
			obj.account = acc.Detail.(*Account)

			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	node.children = make(map[string]*Node)
	node.parent = core.nodes[node.ParentID]
	node.parent.children[node.Name] = node
	core.nodes[node.ID] = node

	return &Message{Type: MessageOK}
}

func updateHandler(msg *Message) (resp *Message) {
	var err error
	if msg.Path == "" {
		return errorMessage(http.StatusBadRequest, "can't update node without a path")
	}
	node := find(core.root, msg.Path, msg.User)
	if node == nil {
		return errorMessage(http.StatusBadRequest, "node doesn't exists")
	}
	err = core.db.Transaction(func(tx *gorm.DB) error {
		switch obj := msg.Payload.(type) {
		case *Broker:
			obj.NodeID = node.ID
			obj.account = node.Detail.(*Broker).account
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Account:
			obj.APIPublicKey, err = encryptString(core.dbKey, node.Name, obj.APIPublicKey)
			if err != nil {
				return err
			}
			obj.APIPrivateKey, err = encryptString(core.dbKey, node.Name, obj.APIPrivateKey)
			if err != nil {
				return err
			}
			obj.NodeID = node.ID
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Group:
			obj.NodeID = node.ID
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		case *Pocket:
			obj.NodeID = node.ID
			if err := tx.Create(&obj).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	node.Detail = msg.Payload.(Detailer)
	core.sendUpdates(node.ID)
	return &Message{Type: MessageOK}
}

func getTreeHandler(msg *Message) (resp *Message) {
	var err error
	resp = &Message{Type: MessageTree}

	resp.Payload, err = json.Marshal(core.nodes[msg.Payload.(uint)].treeMap())
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	return
}

package server

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type messageHandler func(*Core, *Message) *Message

var messageHandlers = map[string]messageHandler{
	MessageCreateAccount:    createAccountHandler,
	MessageGetState:         getStateHandler,
	MessageAuthenticateUser: authenticateUserHandler,
	MessageCreateUser:       createUserHandler,
}

func createAccountHandler(core *Core, msg *Message) (resp *Message) {
	var err error
	aws := msg.Payload.(*AccountWithSecret)

	var accExists bool
	err = core.db.Model(aws.Account).
		Select("count(*)>0").
		Where("user_id = ? AND name = ?", aws.Account.UserID, aws.Account.Name).
		First(&accExists).Error
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	if accExists {
		return errorMessage(http.StatusBadRequest, "account already exists")
	}

	aws.Secret.ApiPublicKey, err = encryptString(core.dbKey, aws.Account.Name, aws.Secret.ApiPublicKey)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	aws.Secret.ApiPrivateKey, err = encryptString(core.dbKey, aws.Account.Name, aws.Secret.ApiPrivateKey)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}

	tx := core.db.Create(aws.Account)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}
	aws.Secret.AccountID = aws.Account.ID
	tx = core.db.Create(aws.Secret)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}

	core.userMap[aws.Account.UserID].Accounts = append(core.userMap[aws.Account.UserID].Accounts, aws.Account)
	core.accountMap[aws.Account.ID] = aws.Account

	// resp.Payload = aws
	// core.broadcastMessage(resp)
	return &Message{Type: MessageOK}
}

func getStateHandler(c *Core, msg *Message) (resp *Message) {
	var err error
	resp = &Message{Type: MessageState}

	resp.JSPayload, err = json.Marshal(c.userMap[msg.Payload.(uint)])
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	return
}

func authenticateUserHandler(core *Core, msg *Message) (resp *Message) {
	uws := msg.Payload.(UserWithSecret)
	exuws := &UserWithSecret{}
	tx := core.db.Where("email = ?", uws.User.Email).First(&exuws.User)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}
	if exuws.User.ID == 0 {
		return errorMessage(http.StatusBadRequest, "user does not exist")
	}

	tx = core.db.Where("user_id = ?", exuws.User.ID).First(&exuws.Secret)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}
	if exuws.Secret.UserID == 0 {
		return errorMessage(http.StatusInternalServerError, "user secret does not exist")
	}
	log.Debug(uws.Secret.Password, exuws.Secret.Password)
	err := compareHashPassword(uws.Secret.Password, exuws.Secret.Password)
	if err != nil {
		return errorMessage(http.StatusBadRequest, err.Error())
	}

	return &Message{Type: MessageUserWithSecret, Payload: exuws.User}
}

func createUserHandler(core *Core, msg *Message) (resp *Message) {
	uws := msg.Payload.(UserWithSecret)
	var userExists bool
	err := core.db.Model(UserDetail{}).Select("count(*)>0").Where("email = ?", uws.User.Email).First(&userExists).Error
	if err != nil {
		return errorMessage(http.StatusInternalServerError, err.Error())
	}
	if userExists {
		return errorMessage(http.StatusBadRequest, "user already exists")
	}

	log.Debug(uws.Secret.Password)
	uws.Secret.Password, err = hashPassword(uws.Secret.Password)
	if err != nil {
		return errorMessage(http.StatusInternalServerError, "could not generate password hash")
	}
	uws.User.Role = RoleUser

	// Add user to database and core
	tx := core.db.Create(uws.User)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}
	uws.Secret.UserID = uws.User.ID
	tx = core.db.Create(uws.Secret)
	if tx.Error != nil {
		return errorMessage(http.StatusInternalServerError, tx.Error.Error())
	}
	core.users = append(core.users, uws.User)
	core.userMap[uws.User.ID] = uws.User

	return &Message{Type: MessageOK}
}

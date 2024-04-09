package server

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	AuthLogin  = "/login"
	AuthSignup = "/signup"
)

func (core *Core) serveHTTP() {
	log.Info("Starting webserver.")

	r := gin.Default()

	// Load GUI templates
	r.LoadHTMLGlob("./web/gui/*")

	// Authentication routes
	r.StaticFile(AuthLogin, "./web/login.html")
	r.StaticFile(AuthSignup, "./web/signup.html")
	r.POST(AuthLogin, login)
	r.POST(AuthSignup, signup)

	// Statuc content routes
	r.Static("/static", "./web/static")

	// Websocket connection routes
	// r.GET("/socket", core.wsHandler, core.needUserRole)

	// Home page route
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/gui/home")
	})

	// Gui
	r.GET("/gui/*path", needUserRoleOrLogin, gui)

	// API routes
	core.apiRoutes(r)

	err := r.Run(core.settings.HTTPAddr)
	core.message <- &Message{
		Type:    MessageWebServerError,
		Payload: err,
	}
}

type LoginResp struct {
	Success bool
	Message string
}

func login(c *gin.Context) {
	user := &User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	resp := core.send(MessageAuthenticateUser, user)
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}
	user = resp.Payload.(*User)

	expirationDuration := ExiprationMins * time.Minute
	expirationTime := time.Now().Add(expirationDuration)
	claims := &Claims{
		Role: user.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.Itoa(int(user.NodeID)),
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(core.jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIError{Error: "could not generate token"})
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("token", tokenString, int(expirationDuration.Seconds()), "/", "localhost", false, true)
	c.JSON(http.StatusOK, &LoginResp{Success: true, Message: "user logged in"})
}

func signup(c *gin.Context) {
	userNode := &Node{
		Detail: &User{},
	}
	if err := c.ShouldBindJSON(userNode); err != nil {
		c.JSON(http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}

	resp := core.send(MessageCreateUser, userNode)
	if resp.Type == MessageError {
		c.JSON(resp.extractError())
		return
	}

	c.JSON(200, gin.H{"success": "user created"})
}

func needUserRole(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	claims, err := core.ParseToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	if claims.Role != RoleUser {
		c.JSON(401, gin.H{"error": "needs user role"})
		c.Abort()
		return
	}

	c.Set("role", claims.Role)
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		c.JSON(401, gin.H{"error": "can't convert user id"})
		c.Abort()
		return
	}
	c.Set("userID", uint(userID))
}

func needUserRoleOrLogin(c *gin.Context) {
	dest := url.Values{}
	dest.Set("dest", c.Param("path"))
	loginURL := "/login?" + dest.Encode()

	token, err := c.Cookie("token")
	if err != nil {
		c.Abort()
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	claims, err := core.ParseToken(token)
	if err != nil {
		c.Abort()
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	if claims.Role != RoleUser {
		c.Abort()
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	c.Set("role", claims.Role)
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		c.JSON(401, gin.H{"error": "can't convert user id"})
		c.Abort()
		return
	}
	c.Set("userID", uint(userID))
}

func gui(c *gin.Context) {
	userID := c.GetUint("userID")
	log.Printf("UserID: %v", userID)
	log.Printf("Nodes: %#v", core.nodes)
	c.HTML(http.StatusOK, "gui.html", core.nodes[userID])
}

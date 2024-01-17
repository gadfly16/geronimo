package server

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (core *Core) serveHTTP() {
	log.Info("Starting webserver.")

	r := gin.Default()

	r.POST("/login", core.login)
	r.POST("/signup", core.signup)
	r.Static("/static", "./static")
	r.GET("/socket", core.wsHandler, core.needUserRole)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static")
	})

	api := r.Group("/api", core.needUserRole)
	{
		api.GET("/full-state", core.getFullState)
	}

	err := r.Run(core.settings.HTTPAddr)
	core.message <- &Message{
		Type:    mt.WebServerError,
		Payload: err,
	}
}

func (core *Core) login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var existingUser User
	core.db.Where("email = ?", user.Email).First(&existingUser)
	if existingUser.ID == 0 {
		c.JSON(400, gin.H{"error": "user does not exist"})
		return
	}

	err := compareHashPassword(user.Password, existingUser.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid password"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Role: existingUser.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   existingUser.Email,
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(core.jwtKey)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", tokenString, int(expirationTime.Unix()), "/", "localhost", false, true)
	c.JSON(200, gin.H{"success": "user logged in"})
}

func (core *Core) signup(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var existingUser User
	core.db.Where("email = ?", user.Email).First(&existingUser)
	if existingUser.ID != 0 {
		c.JSON(400, gin.H{"error": "user already exists"})
		return
	}

	var err error
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate password hash"})
		return
	}

	user.Role = UserRole
	core.db.Create(&user)

	c.JSON(200, gin.H{"success": "user created"})
}

func (core *Core) needUserRole(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	claims, err := core.parseToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}

	c.Set("role", claims.Role)
}

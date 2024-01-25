package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	mt "github.com/gadfly16/geronimo/messagetypes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (core *Core) serveHTTP() {
	log.Info("Starting webserver.")

	r := gin.Default()

	// Load GUI templates
	r.LoadHTMLGlob("./web/gui/*")

	// Authentication routes
	r.StaticFile("/login", "./web/login.html")
	r.StaticFile("/signup", "./web/signup.html")
	r.POST("/login", core.login)
	r.POST("/signup", core.signup)

	// Statuc content routes
	r.Static("/static", "./web/static")

	// Websocket connection routes
	r.GET("/socket", core.wsHandler, core.needUserRole)

	// Home page route
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/gui/home")
	})

	// Gui page routes
	gui := r.Group("/gui", core.needUserRoleOrLogin)
	{
		gui.GET("/user/:id", core.user)
	}

	// API routes
	api := r.Group("/api", core.needUserRole)
	{
		api.GET("/account", core.getAccounts)
		api.POST("/account", core.postAccount)
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

	expirationDuration := 5 * time.Minute
	expirationTime := time.Now().Add(expirationDuration)
	claims := &Claims{
		Role: existingUser.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.Itoa(int(existingUser.ID)),
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(core.jwtKey)

	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("token", tokenString, int(expirationDuration.Seconds()), "/", "localhost", false, true)
	c.JSON(200, gin.H{"success": "user logged in", "userid": existingUser.ID})
}

func (core *Core) signup(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var userExists bool
	err := core.db.Model(&user).Select("count(*)>0").Where("email = ?", user.Email).First(&userExists).Error
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if userExists {
		c.JSON(400, gin.H{"error": "user already exists"})
		return
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate password hash"})
		return
	}

	user.Role = UserRole

	// Add user to database and core
	core.db.Create(&user)
	core.users = append(core.users, &user)
	core.userMap[user.ID] = &user

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

	if claims.Role != UserRole {
		c.JSON(401, gin.H{"error": "needs user role"})
		c.Abort()
		return
	}
	c.Set("role", claims.Role)
	c.Set("userID", claims.Subject)
}

func (core *Core) needUserRoleOrLogin(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		// c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	claims, err := core.parseToken(token)
	if err != nil {
		// c.JSON(401, gin.H{"error": "unauthorized"})
		c.Abort()
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	if claims.Role != UserRole {
		// c.JSON(401, gin.H{"error": "needs user role"})
		c.Abort()
		c.Redirect(http.StatusMovedPermanently, "/login")
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

func (core *Core) user(c *gin.Context) {
	authUserID, _ := c.Get("userID")
	reqUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(401, gin.H{"error": "bad user id"})
		return
	}
	if authUserID.(uint) != uint(reqUserID) {
		c.JSON(401, gin.H{"error": "user has no permission to see data"})
		return
	}
	user := core.userMap[uint(reqUserID)]
	c.HTML(http.StatusOK, "user.html", user)
}

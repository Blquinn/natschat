package main

import (
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"playground/natschat/config"
	"playground/natschat/models"
	"playground/natschat/services"
	"playground/natschat/utils/auth"
)

var (
	addr = flag.String("addr", ":5000", "http service address")
	debug = flag.Bool("debug", false, "use debug mode")

	validate *validator.Validate
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	validate = validator.New()

	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	gnats := newGnats(ns)

	hub := newHub()
	go hub.run()

	log.Println("listening on " + *addr)

	db := config.GetDBConn()
	userService := services.NewUserService(db)

	r := setupRouter(hub, gnats, userService)
	err = r.Run(*addr)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func setupRouter(hub *Hub, gnats *Gnats, userService services.IUserService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if *debug {
		r.Use(gin.Logger())
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowAllOrigins = true
		r.Use(cors.New(corsCfg))
	}

	r.GET("/", func(c *gin.Context) {
		c.File("home.html")
	})
	r.POST("/login", loginHandler(userService))

	api := r.Group("/api")
	api.Use(auth.AuthenticateUserJWT)
	api.GET("/something", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"detail": "Cool!"})
	})

	r.GET("/ws", func (c *gin.Context) {
		serveWs(hub, gnats, c.Writer, c.Request)
	})
	return r
}

func loginHandler(userService services.IUserService) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		var body models.LoginRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		user, err := userService.GetUserByUsername(body.Username)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"detail": "User not found."})
			return
		}

		if user.Password != body.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password."})
			return
		}

		jwt, err := auth.CreateJWT(user)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"detail": "Error occurred while processing request"})
			return
		}

		c.JSON(http.StatusOK, &map[string]string{
			"token": jwt,
		})
	}
}

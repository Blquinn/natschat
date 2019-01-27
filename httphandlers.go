package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"log"
	"natschat/config"
	"natschat/models"
	"natschat/services"
	"natschat/utils"
	"natschat/utils/auth"
	"net/http"
)

type Server struct {
	config      *config.Config
	jwt         *auth.JWT
	db          *gorm.DB
	hub         *Hub
	gnats       *Gnats
	userService services.IUserService
	chatService services.IChatService
}

func (s *Server) Run() {
	r := gin.New()
	r.Use(gin.Recovery())
	if s.config.Debug {
		r.Use(gin.Logger())
	}

	if s.config.Environment == config.EnvironmentLocal {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowAllOrigins = true
		corsCfg.AddAllowHeaders("Authorization")
		r.Use(cors.New(corsCfg))
	}

	r.POST("/register", s.registerUser)
	r.POST("/login", s.loginHandler)

	api := r.Group("/api")
	api.Use(s.jwt.AuthenticateUserJWT)
	api.POST("/rooms", s.createChatRoomHandler)
	api.GET("/rooms", s.listChatRoomsHandler)
	api.GET("/rooms/:id/history", s.chatHistoryHandler)

	r.GET("/ws", func(c *gin.Context) {
		s.serveWs(c.Writer, c.Request)
	})
	err := r.Run(s.config.Server.Address)
	if err != nil {
		logrus.Fatalf("Server died with error: %v", err)
	}
}

// serveWs handles websocket requests from the peer.
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(s.hub, s.gnats, conn, s.chatService)
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (s *Server) registerUser(c *gin.Context) {
	var ur models.CreateUserRequest
	if err := c.ShouldBindJSON(&ur); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	u, err := s.userService.CreateUser(ur)
	if err != nil {
		err.HandleResponse(c)
		return
	}

	c.JSON(http.StatusCreated, u.ToDTO())
}

func (s *Server) loginHandler(c *gin.Context) {
	var body models.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	user, err := s.userService.GetUserByUsername(body.Username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"detail": "User not found."})
		return
	}

	if user.Password != body.Password {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "Incorrect username or password."})
		return
	}

	jwt, err := s.jwt.CreateJWT(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"detail": "Error occurred while processing request"})
		return
	}

	c.JSON(http.StatusOK, &map[string]string{
		"token": jwt,
	})
}

func (s *Server) createChatRoomHandler(c *gin.Context) {
	body := models.CreateChatRoomRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	u := auth.GetUserOrPanic(c)
	r, err := s.chatService.CreateChatRoom(body.Name, u.ID)
	if err != nil {
		err.HandleResponse(c)
		return
	}

	c.JSON(http.StatusCreated, r)
}

func (s *Server) listChatRoomsHandler(c *gin.Context) {
	rooms, err := s.chatService.ListChatRooms()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.ResponseJSON())
		return
	}

	c.JSON(http.StatusOK, models.PageResponse{Results: rooms})
}

func (s *Server) chatHistoryHandler(c *gin.Context) {
	id := c.Param("id")
	msgs, err := s.chatService.GetChatHistory(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.ResponseJSON())
		return
	}

	c.JSON(200, models.PageResponse{Results: msgs})
}

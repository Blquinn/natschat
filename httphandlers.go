package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"time"

	//"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"natschat/components/chat"
	"natschat/components/users"
	"natschat/config"
	"natschat/utils/apierrs"
	"natschat/utils/auth"
	"natschat/utils/pagination"
	"net/http"
)

type Server struct {
	config      *config.Config
	jwt         *auth.JWT
	db          *gorm.DB
	hub         *Hub
	gnats       *Gnats
	userService *users.Service
	chatService *chat.Service
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
		log.Fatalf("Server died with error: %v", err)
	}
}

// serveWs handles websocket requests from the peer.
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	u, err := authenticateSocket(conn, s.jwt)
	if err != nil {
		if err1 := conn.WriteJSON(Message{
			Type: MessageTypeUnauthorizedErr,
			Body: ServerErrorMessage{
				Message: "Auth error occurred",
			},
		}); err != nil {
			log.Errorf("Got err while sending auth message: %v", err1)
		}
		if err2 := conn.Close(); err2 != nil {
			log.Errorf("Got err while closing socket during auth: %v", err2)
		}
		return
	}

	m := Message{
		Type: MessageTypeAuthAck,
		Body: "Authentication success",
	}
	if err := conn.WriteJSON(&m); err != nil {
		log.Errorf("Err while authack %v", err)
		if err2 := conn.Close(); err2 != nil {
			log.Errorf("Got err while closing socket during auth: %v", err2)
		}
		return
	}

	client := newClient(s.hub, s.gnats, conn, s.chatService, u)
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

type AuthRequest struct {
	Token string `binding:"required"`
}

func authenticateSocket(conn *websocket.Conn, jwt *auth.JWT) (*auth.JWTUser, error) {
	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Errorf("Error occurred while setting write deadline during auth: %v", err)
		return nil, err
	}

	var r AuthRequest
	if err := conn.ReadJSON(&r); err != nil {
		return nil, err
	}

	j, err := jwt.ParseAndValidateJWT(r.Token)
	if err != nil {
		return nil, err
	}

	return &j, nil
}

func (s *Server) registerUser(c *gin.Context) {
	var ur users.CreateUserRequest
	if err := c.ShouldBindJSON(&ur); err != nil {
		apierrs.HandleValidationError(c, err)
		return
	}

	u, err := s.userService.CreateUser(ur)
	if err != nil {
		err.HandleResponse(c)
		return
	}

	c.JSON(http.StatusCreated, u)
}

func (s *Server) loginHandler(c *gin.Context) {
	var body users.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	var jwt string
	var err *apierrs.APIError
	if jwt, err = s.userService.LoginUser(body); err != nil {
		err.HandleResponse(c)
		return
	}

	c.JSON(http.StatusOK, &map[string]string{
		"Token": jwt,
	})
}

func (s *Server) createChatRoomHandler(c *gin.Context) {
	body := chat.CreateChatRoomRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		apierrs.HandleValidationError(c, err)
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

	c.JSON(http.StatusOK, pagination.PageResponse{Results: rooms})
}

func (s *Server) chatHistoryHandler(c *gin.Context) {
	id := c.Param("id")
	msgs, err := s.chatService.GetChatHistory(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.ResponseJSON())
		return
	}

	c.JSON(200, pagination.PageResponse{Results: msgs})
}

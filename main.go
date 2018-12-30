package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"natschat/config"
	"natschat/services"
	"natschat/utils/auth"
	"os"
)

var (
	configPath = flag.String("config", "", "config.yml path")

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

	cfg, err := getConfig()
	if err != nil {
		log.Fatalf("Error loading cfg: %v", err)
		os.Exit(1)
	}

	if cfg.Debug {
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

	log.Println("listening on " + cfg.Server.Address)

	db := getDBConn(cfg)
	userService := services.NewUserService(db)
	chatService := services.NewChatService(db)
	jwt := auth.NewJWT(cfg)

	s := Server{
		config:      cfg,
		db:          db,
		jwt:         jwt,
		hub:         hub,
		gnats:       gnats,
		userService: userService,
		chatService: chatService,
	}
	s.Run()
}

func getConfig() (*config.Config, error) {
	var cfg *config.Config
	var err error
	if *configPath == "" {
		if cfg, err = config.Parse("config.yml"); err != nil {
			return cfg, err
		}
	} else {
		if cfg, err = config.Parse(*configPath); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func getDBConn(cfg *config.Config) *sqlx.DB {
	var url string
	if cfg.DB.Host == "" {
		url = "host=localhost port=5432 user=ben password=password dbname=chat sslmode=disable"
	} else {
		url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	}
	return sqlx.MustConnect("postgres", url)
}

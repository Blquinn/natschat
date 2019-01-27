package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"natschat/config"
	"natschat/models"
	"natschat/services"
	"natschat/utils/auth"
	"os"
	"time"
)

var (
	configPath      = flag.String("config", "", "config.yml path")
	gnatsConfigPath = flag.String("gnatsdConf", "gnatsd.conf", "gnatsd config.conf path")
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func main() {
	flag.Parse()

	binding.Validator = models.NewDefaultValidator()

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

	gnatsOpts, err := server.ProcessConfigFile(*gnatsConfigPath)
	if err != nil {
		log.Fatalf("Failed to parse gnatsd config: %v", err)
	}

	runGnatsServer(cfg, gnatsOpts)

	ns, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	log.Infoln("Connected to gnatsd")

	gnats := newGnats(ns)

	hub := newHub()
	go hub.run()

	log.Println("listening on " + cfg.Server.Address)

	db, err := setupDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("Error while closing database connection: %v", err)
		}
	}()

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

func setupDB(cfg *config.Config) (*gorm.DB, error) {
	var url string
	if cfg.DB.Host == "" {
		url = "host=localhost port=5432 user=ben password=password dbname=chat sslmode=disable"
	} else {
		url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
	}
	db, err := gorm.Open("postgres", url)
	if err != nil {
		return db, err
	}

	if cfg.DB.MaxIdleConns != 0 {
		db.DB().SetMaxIdleConns(cfg.DB.MaxIdleConns)
	}

	if cfg.DB.MaxOpenConns != 0 {
		db.DB().SetMaxOpenConns(cfg.DB.MaxOpenConns)
	}

	if cfg.Debug {
		db.LogMode(true)
	}

	return db, nil
}

func runGnatsServer(config *config.Config, opts *server.Options) *server.Server {
	//opts := &defaultGnatsOptions
	// Optionally override for individual debugging of tests
	opts.Logtime = true
	opts.NoLog = !config.Gnatsd.Log
	opts.Trace = config.Gnatsd.Trace
	opts.Debug = config.Gnatsd.Debug

	s := server.New(opts)

	if config.Debug {
		s.ConfigureLogger()
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}

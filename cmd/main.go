package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"jira-discord-webhook/internal/handler"
	"jira-discord-webhook/internal/utils"
)

func main() {
	_ = godotenv.Load()

	logLevel := os.Getenv("LOG_LEVEL")
	var level zap.AtomicLevel
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = level
	zapLogger, _ := cfg.Build()
	defer zapLogger.Sync()
	zap.ReplaceGlobals(zapLogger)

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${ip} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     os.Stdout,
	}))
	app.Post("/webhook", handler.WebhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	userMappingPath := os.Getenv("USER_MAPPING_PATH")
	if userMappingPath == "" {
		userMappingPath = "config/user_mapping.toml"
	}
	if err := utils.LoadUserMapping(userMappingPath); err != nil {
		log.Fatalf("failed to load user mapping: %v", err)
	}
	log.Fatal(app.Listen(":" + port))
}

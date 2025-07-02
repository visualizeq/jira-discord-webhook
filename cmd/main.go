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

	appLogFile := "logs/app.log"
	accessLogFile := "logs/access.log"

	zapLogger, _, err := utils.NewZapLoggerWithRotate(appLogFile, level)
	if err != nil {
		log.Fatalf("failed to create zap logger: %v", err)
	}
	defer zapLogger.Sync()
	zap.ReplaceGlobals(zapLogger)

	// Separate rotate writer for access logs
	accessRotateWriter, err := utils.NewRotateWriter(accessLogFile)
	if err != nil {
		log.Fatalf("failed to create access log rotate writer: %v", err)
	}

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${ip} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     accessRotateWriter,
	}))
	app.Post("/webhook", handler.WebhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	userMappingPath := os.Getenv("USER_MAPPING_PATH")
	if userMappingPath == "" {
		userMappingPath = "config/user_mapping.yaml"
	}
	if err := utils.LoadUserMapping(userMappingPath); err != nil {
		log.Fatalf("failed to load user mapping: %v", err)
	}
	log.Fatal(app.Listen(":" + port))
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"nextcloud-spreadsheet-editor/routes"
	"nextcloud-spreadsheet-editor/services"
	"nextcloud-spreadsheet-editor/utils"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func setupLogger() {
	var encoding string
	var encoderCfg zapcore.EncoderConfig

	encoderCfg = zap.NewDevelopmentEncoderConfig()
	encoding = "console"

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLevel := zap.InfoLevel

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       os.Getenv("ENVIRONMENT") == "dev",
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          encoding,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]any{
			"pid": os.Getpid(),
		},
	}
	logger := zap.Must(config.Build())

	zap.ReplaceGlobals(logger)
}

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Failed to load env file")
	}

	setupLogger()

	// setup router
	mux := http.NewServeMux()

	// setup any auth / cors / logging middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// set up telegram bot
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zap.L().Panic("Failed to init new telegram bot", zap.Error(err))
	}

	wh, err := tgbotapi.NewWebhook(fmt.Sprintf("https://1200e6fc0ee1.ngrok-free.app/%s", bot.Token))
	if err != nil {
		zap.L().Panic("Failed to create telegram bot webhook", zap.Error(err))
	}
	if _, err := bot.Request(wh); err != nil {
		zap.L().Panic("Failed to start telegram bot webhook", zap.Error(err))
	}

	// dependencies
	httpClient := utils.HttpClient{}

	dataService := services.NCDataService{
		Http: &httpClient,
	}
	spreadsheetService := services.ExcelerizeSpreadsheetService{}
	telegramService := services.TelegramService{
		Bot: bot,
	}

	// routes
	dataRoutes := routes.DataRoutes{
		DataService:        &dataService,
		SpreadsheetService: &spreadsheetService,
		MessagingService:   &telegramService,
	}

	// register routes
	mux.HandleFunc("/add", dataRoutes.AddValueForCategory)
	mux.HandleFunc(fmt.Sprintf("/%s", bot.Token), dataRoutes.HandleMessage)

	// configure server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	zap.L().Info("Server starting", zap.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

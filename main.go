package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"telegram-spreadsheet-editor/routes"
	"telegram-spreadsheet-editor/services"
	"telegram-spreadsheet-editor/utils"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	BOT_TOKEN_KEY    string = "TELEGRAM_BOT_TOKEN"
	SERVICE_HOST_KEY string = "SERVICE_HOST"
	HOST_KEY         string = "HOST"
	PORT_KEY         string = "PORT"
	LOG_LEVEL_KEY    string = "LOG_LEVEL"
)

func setupLogger() {
	var encoding string
	var encoderCfg zapcore.EncoderConfig

	encoderCfg = zap.NewDevelopmentEncoderConfig()
	encoding = "console"

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel := os.Getenv(LOG_LEVEL_KEY)
	var zapLevel zapcore.Level
	switch logLevel {
	case "Debug":
		zapLevel = zap.DebugLevel
	case "Info":
		zapLevel = zap.InfoLevel
	case "Warning":
		zapLevel = zap.WarnLevel
	case "Error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.WarnLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       os.Getenv("ENVIRONMENT") == "development",
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

	if err := utils.AssertEnvVars(); err != nil {
		zap.L().Error("Required env vars missing", zap.Error(err))
		return
	}

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
	token := os.Getenv(BOT_TOKEN_KEY)
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zap.L().Panic("Failed to init new telegram bot", zap.Error(err))
	}

	serviceHost := os.Getenv(SERVICE_HOST_KEY)

	zap.L().Info("Creating new telegram webhook", zap.String("host", serviceHost))

	// TODO: Update this to use the golang chan to get updates if it works
	wh, err := tgbotapi.NewWebhook(fmt.Sprintf("%s/%s", serviceHost, bot.Token))
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
	valkeyStorageService := services.NewValkeyStorageService()

	// routes
	dataRoutes := routes.DataRoutes{
		DataService:        &dataService,
		SpreadsheetService: &spreadsheetService,
		MessagingService:   &telegramService,
		StorageService:     valkeyStorageService,
	}

	// register routes
	mux.HandleFunc(fmt.Sprintf("/%s", bot.Token), dataRoutes.HandleMessage)

	host := os.Getenv(HOST_KEY)
	port := os.Getenv(PORT_KEY)

	// configure server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
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

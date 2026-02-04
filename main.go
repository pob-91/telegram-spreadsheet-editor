package main

import (
	"log"
	"os"
	"telegram-spreadsheet-editor/model"
	"telegram-spreadsheet-editor/routes"
	"telegram-spreadsheet-editor/services"
	"telegram-spreadsheet-editor/utils"

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

	// set up telegram bot
	token := os.Getenv(BOT_TOKEN_KEY)
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		zap.L().Panic("Failed to init new telegram bot", zap.Error(err))
	}
	bot.Debug = utils.IsDevelopment()

	// delete current webhooks if they exist
	cfg := tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true,
	}
	if _, err := bot.Request(cfg); err != nil {
		zap.L().DPanic("Failed to delete webhook - cannot proceed.", zap.Error(err))
	}

	zap.L().Info("Authorised for account", zap.String("account", bot.Self.UserName))

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

	// listen for channel updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		message := model.Message{
			TelegramMessage: &update,
		}
		dataRoutes.HandleMessage(&message)
	}
}

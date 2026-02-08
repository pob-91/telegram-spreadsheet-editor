package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"telegram-spreadsheet-editor/inputs"
	"telegram-spreadsheet-editor/model"
	"telegram-spreadsheet-editor/routes"
	"telegram-spreadsheet-editor/services"
	"telegram-spreadsheet-editor/utils"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LOG_LEVEL_KEY string = "LOG_LEVEL"
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
	wd, _ := os.Getwd()
	zap.L().Info("Starting edtitor", zap.String("working dir", wd))

	logLevel := os.Getenv(LOG_LEVEL_KEY)
	if logLevel == "Debug" {
		entries, _ := os.ReadDir("./")
		files := make([]string, len(entries))
		for i, e := range entries {
			if e.IsDir() {
				continue
			}
			files[i] = e.Name()
		}
		zap.L().Debug("Contents of directory", zap.String("contents", strings.Join(files, ", ")))
	}

	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Failed to load env file", zap.Error(err))
	}

	setupLogger()

	if err := utils.AssertEnvVars(); err != nil {
		zap.L().Error("Required env vars missing", zap.Error(err))
		return
	}

	// load config
	config_path := os.Getenv(utils.CONFIG_PATH_KEY)
	config, err := model.NewConfigFromFile(config_path)
	if err != nil {
		zap.L().Panic("Could not load config - cannot proceed", zap.Error(err))
	}
	model.RegisterConfig(config)

	// dependencies
	httpClient := utils.HttpClient{}

	dataService := services.NCDataService{
		Http: &httpClient,
	}
	spreadsheetService := services.ExcelerizeSpreadsheetService{}
	telegramService := services.TelegramService{}
	valkeyStorageService := services.NewValkeyStorageService()

	// routes
	dataRoutes := routes.DataRoutes{
		DataService:        &dataService,
		SpreadsheetService: &spreadsheetService,
		MessagingService:   &telegramService,
		StorageService:     valkeyStorageService,
	}

	// create input handlers for each user's inputs
	inputHandlers := []inputs.Input{}
	for _, u := range config.Users {
		for _, i := range u.Inputs {
			switch i.GetType() {
			case model.INPUT_TYPE_TELEGRAM:
				ti, ok := i.(*model.TelegramInput)
				if !ok {
					zap.L().DPanic("For some reason the telegram input is not *TelegramInput")
					break
				}
				in, err := inputs.NewTelegramInput(ti, u.Name)
				if err != nil {
					// error logs in NewTelegramInput
					break
				}
				inputHandlers = append(inputHandlers, in)
			default:
				zap.L().Error("Unhandled input type", zap.String("type", i.GetType()))
			}
		}
	}

	// start each
	for _, h := range inputHandlers {
		go h.Start(dataRoutes.HandleMessage)
	}

	// listen for shutdown signal
	zap.L().Info("Listening for termination messages SIGINT & SIGTERM")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	for s := range shutdown {
		zap.L().Info("Shutting down", zap.String("signal", s.String()))
		for _, h := range inputHandlers {
			h.Stop()
		}
		close(shutdown)
	}
}

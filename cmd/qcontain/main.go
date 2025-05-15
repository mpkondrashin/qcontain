package main

import (
	"os"
	"path/filepath"

	"qcontain/internal/config"
	"qcontain/internal/contain"
	"qcontain/internal/qcservice"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yeka/zip"
	"gopkg.in/natefinch/lumberjack.v2"
)

func LoadConfig() *config.Config {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	configPath := filepath.Join(filepath.Dir(execPath), config.FileName)
	cfg, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}
	return cfg
}

func ConfigureLogging(cfg *config.LoggingConfig) *logrus.Logger {
	logFilePath := cfg.File
	if !filepath.IsAbs(logFilePath) {
		execPath, err := os.Executable()
		if err != nil {
			panic(err)
		}
		logFilePath = filepath.Join(filepath.Dir(execPath), logFilePath)
	}
	logger := logrus.New()
	logger.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level in config, using 'info'")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	return logger
}

func FabricContain(cfg *config.TargetConfig, logger *logrus.Logger) *contain.Contain {
	contain := contain.NewContain(cfg.Folder, logger)
	contain.SetPassword(cfg.Password)
	contain.SetEncryption(zip.EncryptionMethod(cfg.Encryption))
	return contain
}

func main() {
	install := NewInstall(logrus.New())
	err := install.Install()
	if err != nil {
		if err == ErrInstalled {
			return
		}
		logrus.Error(err)
		return
	}

	cfg := LoadConfig()
	logger := ConfigureLogging(&cfg.Logging)
	contain := FabricContain(&cfg.Target, logger)

	svc := qcservice.New(&cfg.Monitor, contain, logger)
	svcConfig := &service.Config{
		Name:        "qcontain",
		DisplayName: "Quarantine Contain",
		Description: "Quarantine SProtect Quarantine",
	}

	s, err := service.New(svc, svcConfig)
	if err != nil {
		logger.Error(err)
		return
	}
	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			logger.Error(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

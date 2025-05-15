package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"qcontain/internal/config"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type Install struct {
	logger *logrus.Logger
}

func NewInstall(logger *logrus.Logger) *Install {
	return &Install{
		logger: logger,
	}
}

var ErrInstalled = errors.New("installed")

func (i *Install) Install() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir == config.InstallationFolder {
		return nil
	}
	i.logger.Info("Installing to ", config.InstallationFolder)

	upgrade := false
	executable := filepath.Base(path)
	executablePath := filepath.Join(config.InstallationFolder, executable)
	_, err = os.Stat(executablePath)
	if !errors.Is(err, os.ErrNotExist) {
		upgrade = true
		i.logger.Info("Stopping service")
		cmd := exec.Command(executablePath, "stop")
		err = cmd.Run()
		if err != nil {
			return err
		}
		i.logger.Info("Service stopped")
	}
	conf := config.DefaultConfig
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter quarantine folder [%s]: ", conf.Monitor.Folder)
	folder, _ := reader.ReadString('\n')
	folder = strings.TrimSpace(folder)
	if folder != "" {
		conf.Monitor.Folder = folder
	}

	fmt.Printf("Enter password [%s]: ", conf.Target.Password)
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password != "" {
		conf.Target.Password = password
	}

	i.logger.Info("Installing to ", config.InstallationFolder)
	err = os.MkdirAll(config.InstallationFolder, 0755)
	if err != nil {
		return err
	}
	i.logger.Info("Copying executable")
	err = Copy(executablePath, path)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		err = os.Chmod(executablePath, 0755)
		if err != nil {
			return err
		}
	}
	i.logger.Info("Saving config")
	err = config.Save(filepath.Join(config.InstallationFolder, config.FileName), &conf)
	if err != nil {
		return err
	}
	if !upgrade {
		i.logger.Info("Installing service")
		cmd := exec.Command(executablePath, "install")
		err = cmd.Run()
		if err != nil {
			i.logger.WithError(err).Warn(executablePath, "install")
			return err
		}
	}
	i.logger.Info("Starting service")
	cmd := exec.Command(executablePath, "start")
	err = cmd.Run()
	if err != nil {
		return err
	}
	i.logger.Info("Service started")

	fmt.Printf("Press Enter to exist setup")
	reader.ReadString('\n')
	return ErrInstalled
}

func Copy(dstFilePath, srcFilePath string) error {
	src, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

package qcservice

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"qcontain/internal/config"
	"qcontain/internal/contain"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/rjeczalik/notify"
	"github.com/sirupsen/logrus"
)

const DefaultNotifyLength = 10

type QCService struct {
	config  *config.MonitorConfig
	logger  *logrus.Logger
	contain *contain.Contain
	exit    chan struct{}
}

func New(cfg *config.MonitorConfig, contain *contain.Contain, logger *logrus.Logger) *QCService {
	return &QCService{
		config:  cfg,
		logger:  logger,
		contain: contain,
		exit:    make(chan struct{}),
	}
}

func (q *QCService) Start(s service.Service) error {
	q.logger.Info("Start")
	go func() {
		err := q.Run()
		if err != nil {
			q.logger.WithError(err).Error("Service aborted")
		}
	}()
	return nil
}

func (q *QCService) Stop(s service.Service) error {
	q.logger.Info("Stop")
	q.exit <- struct{}{}
	return nil
}

func (s *QCService) Run() error {
	err := s.WaitForSourceFolder()
	if err != nil {
		return err
	}
	s.logger.Info("Prescan")
	if err := s.contain.ProcessFolder(s.config.Folder); err != nil {
		s.logger.WithError(err).Error("Prescan failed")
	}
	s.logger.Info("Service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if s.config.NotifyLength == 0 {
		s.config.NotifyLength = DefaultNotifyLength
	}
	c := make(chan notify.EventInfo, s.config.NotifyLength)

	if err := notify.Watch(s.config.Folder, c, notify.All); err != nil {
		return err
	}
	defer notify.Stop(c)
	s.logger.Info("Monitoring: ", s.config.Folder)
	for {
		select {
		case sig := <-sigChan:
			s.logger.WithField("signal", sig).Info("Received shutdown signal")
			return nil
		case <-s.exit:
			s.logger.Info("Service stopping")
			return nil
		case event := <-c:
			s.logger.Debug("Got event:", event)
			if err := s.ProcessEvent(event); err != nil {
				s.logger.WithError(err).Warn("Failed to process event")
			}
		}
	}
}

func (s *QCService) ProcessEvent(event notify.EventInfo) error {
	s.logger.Debug("Executing service work")
	if event.Event() != notify.Create {
		return nil
	}
	return s.contain.Process(event.Path())
}

var ErrNotDirectory = errors.New("quarantine folder is not a directory")

func (s *QCService) WaitForSourceFolder() error {
	for {
		info, err := os.Stat(s.config.Folder)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("%w: %s", ErrNotDirectory, s.config.Folder)
			}
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		s.logger.Warn(s.config.Folder, ": Quarantine folder not found, waiting...")
		time.Sleep(10 * time.Second)
	}
}

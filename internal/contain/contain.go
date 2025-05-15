package contain

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yeka/zip"
)

type Contain struct {
	logger       *logrus.Logger
	targetFolder string
	password     string
	encryption   zip.EncryptionMethod
}

func NewContain(targetFolder string, logger *logrus.Logger) *Contain {
	return &Contain{
		targetFolder: targetFolder,
		password:     "",
		encryption:   zip.StandardEncryption,
		logger:       logger,
	}
}

func (c *Contain) SetPassword(password string) *Contain {
	c.password = password
	return c
}

func (c *Contain) SetEncryption(encryption zip.EncryptionMethod) *Contain {
	c.encryption = encryption
	return c
}

func (c *Contain) encrypt(filePath string) error {
	if err := os.MkdirAll(c.targetFolder, 0755); err != nil {
		return err
	}
	fileName := filepath.Base(filePath)
	zipFilePath := filepath.Join(c.targetFolder, fileName+".zip")
	fZip, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer func() {
		err := fZip.Close()
		if err != nil {
			if c.logger != nil {
				c.logger.WithError(err).Warn("Failed to close zip file")
			}
		}
	}()
	zipW := zip.NewWriter(fZip)
	defer func() {
		err := zipW.Close()
		if err != nil {
			if c.logger != nil {
				c.logger.WithError(err).Warn("Failed to close zip writer")
			}
		}
	}()
	encryption := zip.EncryptionMethod(c.encryption)
	f, err := zipW.Encrypt(fileName, c.password, encryption)
	if err != nil {
		return err
	}
	inf, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer func() {
		err := inf.Close()
		if err != nil {
			if c.logger != nil {
				c.logger.WithError(err).Warn("Failed to close file")
			}
		}
	}()
	_, err = io.Copy(f, inf)
	return err
}

func (c *Contain) Process(filePath string) error {
	c.logger.Info("Contain file: ", filepath.Base(filePath))
	c.encrypt(filePath)
	return os.Remove(filePath)
}

func (c *Contain) ProcessFolder(folderPath string) error {
	c.logger.Info("Processing folder: ", folderPath)
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if err := c.Process(path); err != nil {
			c.logger.WithError(err).Warn("Failed to process file")
		}
		return nil
	})
}

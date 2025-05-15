package contain

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/yeka/zip"
)

func TestContain(t *testing.T) {
	contain := NewContain("testing", logrus.New())
	contain.SetPassword("test")
	contain.SetEncryption(zip.StandardEncryption)
	fileContent := "test"
	os.WriteFile("test", []byte(fileContent), 0644)
	contain.Process("test")
}

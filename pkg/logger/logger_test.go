package logger

import (
	"testing"

	"github.com/nhirsama/onePushBot/pkg/domain"
)

func TestLogger(t *testing.T) {
	log := NewLog(domain.DebugLevel)
	log.Debug("Debug Message")
	log.Info("Info Message")
	log.Warn("Warn Message")
	log.Error("Error Message")

	log.SetLevel(domain.InfoLevel)

	log.Debug("Debug Message")
	log.Info("Info Message")
	log.Warn("Warn Message")
	log.Error("Error Message")
}

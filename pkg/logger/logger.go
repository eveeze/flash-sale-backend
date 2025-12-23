package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Format waktu ISO8601
	
	var err error
	Log, err = config.Build()
	if err != nil {
		log.Fatal("Gagal inisialisasi logger zap:", err)
	}
	
	// Ganti global logger
	zap.ReplaceGlobals(Log)
}
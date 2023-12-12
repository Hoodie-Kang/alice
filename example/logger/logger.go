package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	var err error

	config := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string {
		"stdout",
	}
	config.ErrorOutputPaths = []string {
		"stderr",
	}
	logger, err = config.Build(zap.AddCallerSkip(1))

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
		),
		zap.NewAtomicLevel(),
	)

	logger := zap.New(core)
	logger.WithOptions(zap.AddCallerSkip(1))
	defer logger.Sync()

	if err != nil {
		panic(err)
	}
}

func Error(message string, args map[string]string) {
	var fields []zapcore.Field
	for str, val := range args {
		fields = append(fields, zap.String(str, val))
	} 
	logger.Error(message, fields...)
	os.Exit(1)
}

func Timeout(message string, args map[string]string) {
	var fields []zapcore.Field
	for str, val := range args {
		fields = append(fields, zap.String(str, val))
	} 
	logger.Error(message, fields...)
	os.Exit(250)
}
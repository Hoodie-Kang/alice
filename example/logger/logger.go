package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

func init() {
	var err error

	path := "./tss.log"
	logRotate := &lumberjack.Logger{
		Filename:	path,
		MaxSize:	10,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

	config := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	// config := zap.NewDevelopmentConfig()
	// encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string {
		// "stdout",
		path,
	}
	config.ErrorOutputPaths = []string {
		"stderr",
	}
	logger, err = config.Build(zap.AddCallerSkip(1))

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(logRotate),
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

func Info(message string, args map[string]string) {
	var fields []zapcore.Field
	for str, val := range args {
		fields = append(fields, zap.String(str, val))
	} 
	logger.Info(message, fields...)
}

func Warn(message string, args map[string]string) {
	var fields []zapcore.Field
	for str, val := range args {
		fields = append(fields, zap.String(str, val))
	} 
	logger.Warn(message, fields...)
}

func Error(message string, args map[string]string) {
	var fields []zapcore.Field
	for str, val := range args {
		fields = append(fields, zap.String(str, val))
	} 
	logger.Error(message, fields...)
	os.Exit(1)
}
package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logging *zap.Logger

// Logger Initialization
func init() {
	logFilePath, err := GetModuleDirectoryPath()
	if err != nil {
		log.Fatalf("Error getting file: %s", err.Error())
	}

	// Check for log directory (if not, create)
	logDirectory := fmt.Sprintf("%s/.logs", logFilePath)
	info, err := ensureDirectory(logDirectory)
	if err != nil {
		log.Fatalf("Error creating log directory: %s", err.Error())
	}

	// Create log file based on current date (if exissts, append)
	logPath := fmt.Sprintf("%s/log_%s.log", logDirectory, time.Now().Format("2006-01-02"))
	if info != nil {
		logPath = fmt.Sprintf("%s/log_%s.log", logDirectory, info.ModTime().Format("2006-01-02"))
	}

	if err := ensureLogFile(logPath); err != nil {
		log.Fatalf("Error creating log file: %s", err.Error())
	}

	// Custom log configuration
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     365, // days
			Compress:   false,
			LocalTime:  true,
		}),
		zapcore.DebugLevel,
	)

	// Zap logger core setup
	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()
	Logging = logger
}

// Create log directory
func ensureDirectory(directoryPath string) (os.FileInfo, error) {
	info, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directoryPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
		info, err = os.Stat(directoryPath)
	}

	return info, err
}

// Create log file
func ensureLogFile(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create log file: %v", err)
		}
		file.Close()
	}
	return nil
}

// // HTTP Log Interceptor with internal logger and log file
// func HTTPLogInterceptor(logger *zap.Logger) gin.HandlerFunc {
// 	fmt.Println("HTTPLogInterceptor Called ✅✅")

// 	return func(c *gin.Context) {
// 		// Start time of request
// 		start := time.Now()
// 		// Let request process
// 		c.Next()
// 		// End time of request
// 		end := time.Now()

// 		// Log details using the provided logger
// 		logger.Info("HTTP Request",
// 			zap.String("Method", c.Request.Method),
// 			zap.String("Path", c.Request.URL.Path),
// 			zap.Int("Status", c.Writer.Status()),
// 			zap.Duration("Latency", end.Sub(start)),
// 			zap.String("ClientIP", c.ClientIP()),
// 			zap.String("UserAgent", c.Request.UserAgent()),
// 		)
// 	}
// }

func LogDebug(source, activity, debugString string, object ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	Logging.Debug(debugString,
		zap.String("Source", source),
		zap.Any("Object", object),
		zap.String("Activity", activity),
		zap.String("Caller", caller),
	)
}

func LogError(source string, activity string, object interface{}, err error) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	Logging.Error("Error",
		zap.String("Source", source),
		zap.Any("Object", object),
		zap.String("Activity", activity),
		zap.String("Caller", caller),
		zap.Error(err),
	)
}

func LogFatal(source string, activity string, object interface{}, err error) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	Logging.Fatal("Fatal",
		zap.String("Source", source),
		zap.Any("Object", object),
		zap.String("Activity", activity),
		zap.String("Caller", caller),
		zap.Error(err),
	)
}

func LogInfo(source, activity, debugString string, object ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	Logging.Info(debugString,
		zap.String("Source", source),
		zap.Any("Object", object),
		zap.String("Activity", activity),
		zap.String("Caller", caller),
	)
}

func LogWarning(source string, activity string, object interface{}, err ...error) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%d", file, line)
	var errMsg string
	if len(err) > 0 {
		errMsg = err[0].Error()
	}
	Logging.Warn("Warning",
		zap.String("Source", source),
		zap.Any("Object", object),
		zap.String("Activity", activity),
		zap.String("Caller", caller),
		zap.String("Error", errMsg),
	)
}
func GetModuleDirectoryPath() (string, error) {

	executablePath, err := os.Executable()
	if err != nil {
		LogFatal("GetModuleDirectoryPath", "error getting the path of the executable", "", err)
		return "", err
	}
	return filepath.Dir(executablePath), nil
}

package turbo

import (
	logger "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
)

var log *logger.Logger

// ContextHook is a hook to be fired when logging on the logging levels returned from
// `Levels()` on your implementation of the interface. Note that this is not
// fired in a goroutine or a channel with workers, you should handle such
// functionality yourself if your call is non-blocking and you don't wish for
// the logging calls for levels returned from `Levels()` to block.
//
// The original hook interface is:
// type Hook interface {
//	 Levels() []Level
//	 Fire(*Entry) error
// }
type ContextHook struct{}

// Levels returns active log levels
func (hook ContextHook) Levels() []logger.Level {
	return logger.AllLevels
}

// Fire is for adding file, func and line info to logger.
func (hook ContextHook) Fire(entry *logger.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(7, pc)

	for i := 0; i < cnt; i++ {
		pci := pc[i] - 1
		fu := runtime.FuncForPC(pci)
		name := fu.Name()
		if !strings.Contains(name, "github.com/Sirupsen/logrus") {
			file, line := fu.FileLine(pci)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}
	return nil
}

func setupLoggerFile(c *Config) {
	//set up log file.
	logPath := c.configs[turboLogPath]
	if len(strings.TrimSpace(logPath)) == 0 {
		logPath = "log"
	}
	if !path.IsAbs(logPath) && len(strings.TrimSpace(c.ServiceRootPath())) != 0 {
		logPath = c.ServiceRootPathAbsolute() + "/" + logPath
	}
	logPath = path.Clean(logPath)
	if err := os.MkdirAll(logPath, 0755); err == nil {
		logFile := path.Clean(logPath + "/turbo.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		} else {
			panic("Failed to setup log path, please check your service.yaml config.")
		}
	} else {
		panic("Failed to setup log path, please check your service.yaml config.")
	}
}

func initLogger(c *Config) {
	if c.Env() == "production" {
		//init log file path
		setupLoggerFile(c)
		// Log as JSON instead of the default ASCII formatter.
		logger.SetFormatter(&logger.JSONFormatter{})
		//set up log level, info level by default.
		logger.SetLevel(logger.InfoLevel)
	} else {
		logger.SetFormatter(&logger.TextFormatter{})
		logger.SetOutput(os.Stderr)
		// set logger with debug level in development environment.
		logger.SetLevel(logger.DebugLevel)
		logger.AddHook(ContextHook{})
	}

	log = logger.StandardLogger()
}

// SetOutput sets output at runtime
func SetOutput(out io.Writer) {
	log.Out = out
	log.Formatter = &logger.TextFormatter{}
}

package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var hasLoad = false
var lg = logrus.New()

// loadConfig - Load intial config
func loadConfig() {
	lg.SetReportCaller(true)

	f := "/log.log"
	if os.Getenv("log_file") != "" {
		f = "/" + os.Getenv("log_file")
	}

	lg.SetOutput(&lumberjack.Logger{
		Filename:   os.Getenv("dir_log") + f,
		MaxSize:    15, // in megabytes
		MaxBackups: 0,
		MaxAge:     0,    // in days
		Compress:   true, // disabled by default
	})

	hasLoad = true
}

// trace - Backtrace log
func trace(stack int) string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(stack, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	dir := os.Getenv("dir_root")
	file = strings.Replace(file, dir, "", 1)
	return fmt.Sprintf("%s:%d %s\n", file, line, f.Name())
}

// MakeLogEntry - Write log to file and/or print out
func MakeLogEntry(c echo.Context, doTrace bool) *logrus.Entry {
	if !hasLoad {
		loadConfig()
	}
	f := map[string]interface{}{
		"at": time.Now().UTC().Format("2006-01-02 15:04:05"),
	}
	if doTrace {
		f["trace3"] = trace(3)
		f["trace4"] = trace(4)
	}
	if c != nil {
		f["method"] = c.Request().Method
		f["uri"] = c.Request().URL.String()
		f["ip"] = c.Request().RemoteAddr
		f["real_ip"] = c.Request().Header.Get("X-Real-Ip")
		f["proxy_ip"] = c.Request().Header.Get("X-Proxy-Ip")
	}

	return lg.WithFields(f)
}

// PrintLogEntry - Print log entry to stdout
func PrintLogEntry(t string, s string, doTrace bool) {
	fmt.Printf("%s | %s \n", time.Now().UTC().Format("2006-01-02 15:04:05 -0700 MST"), s)
	if t == "info" {
		MakeLogEntry(nil, doTrace).Info(s)
	} else {
		MakeLogEntry(nil, doTrace).Error(s)
	}
}

// EchoLogger -
func EchoLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		MakeLogEntry(c, false).Info("incoming request")
		return next(c)
	}
}

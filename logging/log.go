package logging

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var root = &logrus.Logger{
	Out:   os.Stdout,
	Level: logrus.TraceLevel,
	Formatter: &prefixed.TextFormatter{
		DisableColors: func() bool {
			term, ok := os.LookupEnv("TERM")
			return term == "" || !ok
		}(),
		ForceFormatting: true,
		TimestampFormat: "2006-01-02 15:04:05",
	},
}

type ChildLogger struct {
	parent *logrus.Logger
	prefix string
	level  logrus.Level
}

func NewChildLogger(parent *logrus.Logger, prefix string) *ChildLogger {
	lc := &ChildLogger{
		parent: parent,
		prefix: prefix,
	}

	return lc
}

func (l *ChildLogger) shouldOutput(level logrus.Level) bool {
	return l.level >= level
}

func (l *ChildLogger) Debug(args ...interface{}) {
	if l.shouldOutput(logrus.DebugLevel) {
		l.parent.WithField("prefix", l.prefix).Debug(args...)
	}
}

func (l *ChildLogger) Info(args ...interface{}) {
	if l.shouldOutput(logrus.InfoLevel) {
		l.parent.WithField("prefix", l.prefix).Info(args...)
	}
}

func (l *ChildLogger) Warning(args ...interface{}) {
	if l.shouldOutput(logrus.WarnLevel) {
		l.parent.WithField("prefix", l.prefix).Warning(args...)
	}
}

func (l *ChildLogger) Error(args ...interface{}) {
	if l.shouldOutput(logrus.ErrorLevel) {
		l.parent.WithField("prefix", l.prefix).Error(args...)
	}
}

func (l *ChildLogger) Fatal(args ...interface{}) {
	if l.shouldOutput(logrus.FatalLevel) {
		l.parent.WithField("prefix", l.prefix).Fatal(args...)
	}
}

func (l *ChildLogger) Debugf(format string, args ...interface{}) {
	if l.shouldOutput(logrus.DebugLevel) {
		l.parent.WithField("prefix", l.prefix).Debugf(format, args...)
	}
}

func (l *ChildLogger) Infof(format string, args ...interface{}) {
	if l.shouldOutput(logrus.InfoLevel) {
		l.parent.WithField("prefix", l.prefix).Infof(format, args...)
	}
}

func (l *ChildLogger) Warningf(format string, args ...interface{}) {
	if l.shouldOutput(logrus.WarnLevel) {
		l.parent.WithField("prefix", l.prefix).Warningf(format, args...)
	}
}

func (l *ChildLogger) Errorf(format string, args ...interface{}) {
	if l.shouldOutput(logrus.ErrorLevel) {
		l.parent.WithField("prefix", l.prefix).Errorf(format, args...)
	}
}

func (l *ChildLogger) Fatalf(format string, args ...interface{}) {
	if l.shouldOutput(logrus.FatalLevel) {
		l.parent.WithField("prefix", l.prefix).Fatalf(format, args...)
	}
}

func (l *ChildLogger) IsDebug() bool {
	return l.level >= logrus.DebugLevel
}

func (l *ChildLogger) SetDebug(debug bool) {
	if debug {
		l.level = logrus.DebugLevel
	} else {
		l.level = logrus.InfoLevel
	}
}

type Children struct {
	Main *ChildLogger
	USB  *ChildLogger
	MTP  *ChildLogger
	Data *ChildLogger
	LV   *ChildLogger
}

var log = &Children{
	Main: NewChildLogger(root, "main"),
	USB:  NewChildLogger(root, "usb"),
	MTP:  NewChildLogger(root, "mtp"),
	Data: NewChildLogger(root, "data"),
	LV:   NewChildLogger(root, "lv"),
}

func SetLogLevel(main, usb, mtp, data, lv bool) {
	log.Main.SetDebug(main)
	log.USB.SetDebug(usb)
	log.MTP.SetDebug(mtp)
	log.Data.SetDebug(data)
	log.LV.SetDebug(lv)
}

func GetLogger() *Children {
	return log
}

func HTTPLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			root.WithField("prefix", "http").Infof("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		}()
		next.ServeHTTP(w, r)
	})
}

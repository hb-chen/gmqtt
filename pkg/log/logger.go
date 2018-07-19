package log

import (
	"log"
	"os"

	"github.com/fatih/color"
	l "github.com/smallnest/rpcx/log"
)

const (
	DEBUG Lvl = iota
	INFO
	WARN
	ERROR
	OFF
	fatalLvl
	panicLvl
)

const (
	calldepth = 5
)

var (
	level       = DEBUG
	colorEnable = true
)

func init() {
	logger := &defaultLogger{Logger: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile), calldepth: calldepth}
	l.SetLogger(logger)
}

type (
	Lvl       uint
	colorFunc func(format string, a ...interface{}) string
)

func (lvl Lvl) String() string {
	switch lvl {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return lvl.colorString("INFO", color.GreenString)
	case WARN:
		return lvl.colorString("WARN", color.YellowString)
	case ERROR:
		return lvl.colorString("ERROR", color.RedString)
	case fatalLvl:
		return lvl.colorString("FATAL", color.MagentaString)
	case panicLvl:
		return "PANIC"
	default:
		return ""
	}
}

func (lvl Lvl) colorString(str string, f colorFunc) string {
	if colorEnable {
		return f(str)
	} else {
		return str
	}
}

func SetLevel(lvl Lvl) {
	level = lvl
}

func SetColor(enable bool) {
	colorEnable = enable
}

func Debug(v ...interface{}) {
	l.Debug(v...)
}
func Debugf(format string, v ...interface{}) {
	l.Debugf(format, v...)
}

func Info(v ...interface{}) {
	l.Info(v...)
}
func Infof(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func Warn(v ...interface{}) {
	l.Warn(v...)
}
func Warnf(format string, v ...interface{}) {
	l.Warnf(format, v...)
}

func Error(v ...interface{}) {
	l.Error(v...)
}
func Errorf(format string, v ...interface{}) {
	l.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	l.Fatal(v...)
}
func Fatalf(format string, v ...interface{}) {
	l.Fatalf(format, v...)
}

func Panic(v ...interface{}) {
	l.Panic(v...)
}
func Panicf(format string, v ...interface{}) {
	l.Panicf(format, v...)
}

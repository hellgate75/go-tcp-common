package log

import (
	"fmt"
	"github.com/gookit/color"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel string
type LogLevelValue int

const (
	traceLevel LogLevelValue = iota + 1
	debugLevel
	infoLevel
	warningLevel
	errorLevel
	fatalLevel
	TRACE LogLevel = LogLevel("TRACE")
	DEBUG LogLevel = LogLevel("DEBUG")
	INFO  LogLevel = LogLevel("INFO")
	WARN  LogLevel = LogLevel("WARN")
	ERROR LogLevel = LogLevel("ERROR")
	FATAL LogLevel = LogLevel("FATAL")
)

type Logger interface {
	Tracef(format string, in ...interface{})
	Trace(in ...interface{})
	Debugf(format string, in ...interface{})
	Debug(in ...interface{})
	Infof(format string, in ...interface{})
	Info(in ...interface{})
	Warnf(format string, in ...interface{})
	Warn(in ...interface{})
	Errorf(format string, in ...interface{})
	Error(in ...interface{})
	Fatalf(format string, in ...interface{})
	Fatal(in ...interface{})
	Printf(format string, in ...interface{})
	Println(in ...interface{})
	SetVerbosity(verbosity LogLevel)
	GetVerbosity() LogLevel
	Successf(format string, in ...interface{})
	Success(in ...interface{})
	Failuref(format string, in ...interface{})
	Failure(in ...interface{})
}

type logger struct {
	verbosity LogLevelValue
	onScreen  bool
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	prefix    string     // prefix on each line to identify the logger (but see Lmsgprefix)
	flag      int        // properties
	out       io.Writer  // destination for output
	buf       []byte     // for accumulating text to write

}

func (logger *logger) Tracef(format string, in ...interface{}) {
	logger.log(traceLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Trace(in ...interface{}) {
	logger.log(traceLevel, in...)
}

func (logger *logger) Debugf(format string, in ...interface{}) {
	logger.log(debugLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Debug(in ...interface{}) {
	logger.log(debugLevel, in...)
}

func (logger *logger) Infof(format string, in ...interface{}) {
	logger.log(infoLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Info(in ...interface{}) {
	logger.log(infoLevel, in...)
}

func (logger *logger) Warnf(format string, in ...interface{}) {
	logger.log(warningLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Warn(in ...interface{}) {
	logger.log(warningLevel, in...)
}

func (logger *logger) Errorf(format string, in ...interface{}) {
	logger.log(errorLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Error(in ...interface{}) {
	logger.log(errorLevel, in...)
}

func (logger *logger) Fatalf(format string, in ...interface{}) {
	logger.log(fatalLevel, fmt.Sprintf(format, in...))
}

func (logger *logger) Fatal(in ...interface{}) {
	logger.log(fatalLevel, in...)
}

func (logger *logger) Printf(format string, in ...interface{}) {
	var buf []byte = []byte(fmt.Sprintf(format, in...))
	if logger.onScreen {
		color.LightWhite.Printf(string(buf))
	} else {
		logger.out.Write(buf)
	}
}

func (logger *logger) Println(in ...interface{}) {
	var buf []byte = []byte(fmt.Sprint(in...) + "\n")
	if logger.onScreen {
		color.LightWhite.Printf(string(buf))
	} else {
		logger.out.Write(buf)
	}
}

func (logger *logger) SetVerbosity(verbosity LogLevel) {
	logger.verbosity = toVerbosityLevelValue(verbosity)
}
func (logger *logger) GetVerbosity() LogLevel {
	return toVerbosityLevel(logger.verbosity)
}

func (logger *logger) Successf(format string, in ...interface{}) {
	var itfs string = " SUCCESS " + fmt.Sprintf(format, in...) + "\n"
	logger.output(color.Green, 2, itfs)
}

func (logger *logger) Success(in ...interface{}) {
	var itfs string = " SUCCESS " + fmt.Sprint(in...) + "\n"
	logger.output(color.Green, 2, itfs)
}

func (logger *logger) Failuref(format string, in ...interface{}) {
	var itfs string = " FAILURE " + fmt.Sprintf(format, in...) + "\n"
	logger.output(color.Red, 2, itfs)
}

func (logger *logger) Failure(in ...interface{}) {
	var itfs string = " FAILURE " + fmt.Sprint(in...) + "\n"
	logger.output(color.Red, 2, itfs)
}

func (logger *logger) log(level LogLevelValue, in ...interface{}) {
	if level >= logger.verbosity {
		var itfs string = " " + string(toVerbosityLevel(level)) + " " + fmt.Sprint(in...) + "\n"
		switch string(toVerbosityLevel(level)) {
		case "DEBUG":
			logger.output(color.Yellow, 2, itfs)
			break
		case "TRACE":
			logger.output(color.Yellow, 2, itfs)
			break
		case "WARN":
			logger.output(color.LightYellow, 2, itfs)
			break
		case "INFO":
			logger.output(color.LightWhite, 2, itfs)
			break
		case "ERROR":
			logger.output(color.LightRed, 2, itfs)
			break
		case "FATAL":
			logger.output(color.Red, 2, itfs)
			break
		default:
			logger.output(color.White, 2, itfs)
		}
	}
}

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (l *logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	*buf = append(*buf, l.prefix...)
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (logger *logger) output(color color.Color, calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if logger.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		logger.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		logger.mu.Lock()
	}
	logger.buf = logger.buf[:0]
	logger.formatHeader(&logger.buf, now, file, line)
	logger.buf = append(logger.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		logger.buf = append(logger.buf, '\n')
	}
	if logger.onScreen {
		color.Printf(string(logger.buf))
		return nil
	} else {
		_, err := logger.out.Write(logger.buf)
		return err
	}
}

func NewLogger(verbosity LogLevel) Logger {
	return &logger{
		verbosity: toVerbosityLevelValue(verbosity),
		onScreen:  true,
		out:       os.Stdout,
		prefix:    "[go-deploy] ",
		flag:      LstdFlags | LUTC,
	}
}

func NewAppLogger(appName string, verbosity LogLevel) Logger {
	return &logger{
		verbosity: toVerbosityLevelValue(verbosity),
		onScreen:  true,
		out:       os.Stdout,
		prefix:    "[" + appName + "] ",
		flag:      LstdFlags | LUTC,
	}
}

func VerbosityLevelFromString(verbosity string) LogLevel {
	switch strings.ToUpper(verbosity) {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	}
	return INFO
}

func toVerbosityLevelValue(verbosity LogLevel) LogLevelValue {
	switch strings.ToUpper(string(verbosity)) {
	case "TRACE":
		return traceLevel
	case "DEBUG":
		return debugLevel
	case "INFO":
		return infoLevel
	case "WARN":
		return warningLevel
	case "ERROR":
		return errorLevel
	case "FATAL":
		return fatalLevel
	}
	return infoLevel
}

func toVerbosityLevel(verbosity LogLevelValue) LogLevel {
	switch verbosity {
	case traceLevel:
		return TRACE
	case debugLevel:
		return DEBUG
	case infoLevel:
		return INFO
	case warningLevel:
		return WARN
	case errorLevel:
		return ERROR
	case fatalLevel:
		return FATAL
	}
	return INFO
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

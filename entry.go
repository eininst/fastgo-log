package flog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var sDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	sDir = regexp.MustCompile(`utils.utils\.go`).ReplaceAllString(file, "")
}

func entryFileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, sDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

type Fields map[string]interface{}

type Entry struct {
	logger *logger
	data   Fields
}

func NewEntry(logger *logger) *Entry {
	return &Entry{
		logger: logger,
		data:   make(Fields),
	}
}

func (e *Entry) With(fields Fields) *Entry {
	for k, v := range fields {
		e.data[k] = v
	}
	return e
}

func (e *Entry) Trace(a ...any) {
	if e.logger.LogLevel >= TraceLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.traceStr, f, e.fieldsFormat(White), a...))
	}
}

func (e *Entry) Tracef(format string, a ...any) {
	if e.logger.LogLevel >= TraceLevel {
		fmt.Println(e.entrySprintf(e.logger.traceStr, format, e.fieldsFormat(White), a...))
	}
}

func (e *Entry) Debug(a ...any) {
	if e.logger.LogLevel >= DebugLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.debugStr, f, e.fieldsFormat(White), a...))
	}
}

func (e *Entry) Debugf(format string, a ...any) {
	if e.logger.LogLevel >= DebugLevel {
		fmt.Println(e.entrySprintf(e.logger.debugStr, format, e.fieldsFormat(White), a...))
	}
}

func (e *Entry) Info(a ...any) {
	if e.logger.LogLevel >= InfoLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.infoStr, f, e.fieldsFormat(Cyan), a...))
	}
}

func (e *Entry) Infof(format string, a ...any) {
	if e.logger.LogLevel >= InfoLevel {
		fmt.Println(e.entrySprintf(e.logger.infoStr, format, e.fieldsFormat(Cyan), a...))
	}
}

func (e *Entry) Warn(a ...any) {
	if e.logger.LogLevel >= WarnLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.warnStr, f, e.fieldsFormat(Yellow), a...))
	}
}

func (e *Entry) Warnf(format string, a ...any) {
	if e.logger.LogLevel >= WarnLevel {
		fmt.Println(e.entrySprintf(e.logger.warnStr, format, e.fieldsFormat(Yellow), a...))
	}
}

func (e *Entry) Error(a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.errStr, f, e.fieldsFormat(Red), a...))
	}
}

func (e *Entry) Errorf(format string, a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		fmt.Println(e.entrySprintf(e.logger.errStr, format, e.fieldsFormat(Red), a...))
	}
}

func (e *Entry) Fatal(a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		f := e.logger.formatData(a...)
		fmt.Println(e.entrySprintf(e.logger.fatalStr, f, e.fieldsFormat(Red), a...))
		os.Exit(1)
	}
}

func (e *Entry) Fatalf(format string, a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		fmt.Println(e.entrySprintf(e.logger.fatalStr, format, e.fieldsFormat(Red), a...))
		os.Exit(1)
	}
}

func (e *Entry) Panic(a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		f := e.logger.formatData(a...)
		r := e.entrySprintf(e.logger.panicStr, f, e.fieldsFormat(Red), a...)
		fmt.Println(r)
		panic(errors.New(r))
	}
}

func (e *Entry) Panicf(format string, a ...any) {
	if e.logger.LogLevel >= ErrorLevel {
		r := e.entrySprintf(e.logger.panicStr, format, e.fieldsFormat(Red), a...)
		fmt.Println(r)
		panic(errors.New(r))
	}
}

func (e *Entry) getPath() string {
	path := entryFileWithLineNum()
	if !e.logger.FullPath {
		arr := strings.Split(path, "/")
		path = arr[len(arr)-1]
	}
	return path
}

func (e *Entry) entrySprintf(levelStr string, format string, fields any, a ...any) string {
	path := e.getPath()
	msg := fmt.Sprintf(format, a...)
	mlen := len(msg)
	if mlen < e.logger.MsgMinLen {
		for i := 0; i < e.logger.MsgMinLen-mlen; i++ {
			msg += " "
		}
	}
	data := map[string]any{
		"level": levelStr,
		"time":  e.logger.t(),
		"path":  path,
		"msg":   msg,
	}
	if fields != nil && e.logger.writeFields {
		if e.logger.Json {
			for k, v := range fields.(Fields) {
				data[k] = v
			}
			delete(data, "fields")
		} else {
			data["fields"] = fields
		}
	}
	if e.logger.Json {
		s, ok := levelStrMap[data["level"].(string)]
		if ok {
			data["level"] = s
		}
		jsonStr, _ := json.Marshal(data)

		return string(jsonStr)
	} else {
		return Sprintf(e.logger.Format, data)
	}
}

func (e *Entry) fieldsFormat(color string) any {
	if e.logger.Json {
		return e.data
	}
	r := ""
	i := 0

	for k, v := range e.data {
		if i > 0 {
			r += " "
		}
		r += fmt.Sprintf("%s%s%s=%v", color, k, Reset, v)
		i += 1
	}
	return r
}

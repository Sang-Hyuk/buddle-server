package log

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	ErrUndefinedFormatter = fmt.Errorf("undefined Formatter")
)

func Init(c Config) error {
	logrus.SetLevel(c.Level)
	form, err := c.Formatter.Logrus(c.File == nil)
	if err != nil {
		return errors.Wrap(err, "failed to get logrus formatter")
	}
	logrus.SetFormatter(form)
	logrus.SetReportCaller(c.Caller)
	if c.File != nil {
		logrus.SetOutput(c.File.Output())
	}

	return nil
}

type Config struct {
	Level     logrus.Level `yaml:"level"`
	Formatter Formatter    `yaml:"formatter"`
	Caller    bool         `yaml:"caller"`
	File      *File        `yaml:"file,omitempty"`
}

type File struct {
	Path      string `yaml:"path"`
	Hostname  bool   `yaml:"hostname"`
	MaxSize   int    `yaml:"max_size"`
	MaxBackup int    `yaml:"max_backup"`
	MaxAge    int    `yaml:"max_age"`
}

func (f *File) Output() io.Writer {
	return &lumberjack.Logger{
		Filename:   f.formattedName(),
		MaxSize:    f.MaxSize,
		MaxAge:     f.MaxAge,
		MaxBackups: f.MaxBackup,
		LocalTime:  true,
		Compress:   true,
	}
}

func (f *File) formattedName() string {
	if f.Hostname {
		fn := path.Base(f.Path)
		dir := path.Dir(f.Path)
		ext := path.Ext(fn)
		hostname, _ := os.Hostname()
		return path.Join(dir, fmt.Sprintf("%s-%s", strings.TrimSuffix(fn, ext), hostname)+ext)
	}

	return f.Path
}

type Formatter int

const (
	UndefinedFormatter Formatter = iota - 1
	TextFormatter
	JSONFormatter
	PrettyJSONFormatter
)

const (
	textFormatterString       = "text"
	jsonFormatterString       = "json"
	prettyJSONFormatterString = "pretty_json"
)

func (f Formatter) Logrus(colored bool) (logrus.Formatter, error) {
	callerPrettier := func(f *runtime.Frame) (function string, file string) {
		_, filename := path.Split(f.File)
		filename = fmt.Sprintf("%s:%d", filename, f.Line)
		return path.Base(f.Function), filename
	}

	switch f { //nolint:exhaustive
	case TextFormatter:
		return &logrus.TextFormatter{
			ForceColors:      colored,
			CallerPrettyfier: callerPrettier,
			DisableQuote:     true,
		}, nil
	case JSONFormatter:
		return &logrus.JSONFormatter{
			CallerPrettyfier: callerPrettier,
		}, nil
	case PrettyJSONFormatter:
		return &logrus.JSONFormatter{
			CallerPrettyfier: callerPrettier,
			PrettyPrint:      true,
		}, nil
	}

	return nil, ErrUndefinedFormatter
}

func (f *Formatter) UnmarshalText(b []byte) error {
	*f = FormatterFromString(string(b))
	if *f == UndefinedFormatter {
		return ErrUndefinedFormatter
	}

	return nil
}

func (f Formatter) String() string {
	switch f { //nolint:exhaustive
	case TextFormatter:
		return textFormatterString
	case JSONFormatter:
		return jsonFormatterString
	case PrettyJSONFormatter:
		return prettyJSONFormatterString
	}

	return "undefined"
}

func FormatterFromString(str string) Formatter {
	switch str {
	case "", textFormatterString:
		return TextFormatter
	case jsonFormatterString:
		return JSONFormatter
	case prettyJSONFormatterString:
		return PrettyJSONFormatter
	}

	return UndefinedFormatter
}

package logrotator

import (
	"awesomeProject/logging/logrotator/rotationunit"
	"awesomeProject/logging/severity"
	"bytes"
	"context"
	"errors"
	"html/template"
	"log"
	"os"
	"sync"
	"time"
)

type LogRotator struct {
	appName         string
	addr            string
	defaultSeverity severity.Severity
	filePathTmpl    *template.Template
	rotationUnit    rotationunit.Unit

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	current *logger
}

type logger struct {
	FileName         string
	F                *os.File
	UnderlyingLogger *log.Logger
}

type timeData struct {
	Year      int
	YearShort int
	Month     int
	Day       int
	Hour      int
	Minute    int
	Second    int
}

type tmplData struct {
	TimeData timeData
	AppName  string
	Addr     string
}

func New(
	pctx context.Context,
	wg *sync.WaitGroup,
	appName string,
	addr string,
	filePathTmpl string,
	rotationUnit rotationunit.Unit,
	defaultSeverity severity.Severity,
) (*LogRotator, error) {
	tmpl, err := template.New("filePath").Parse(filePathTmpl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(pctx)

	rot := &LogRotator{
		appName:         appName,
		addr:            addr,
		filePathTmpl:    tmpl,
		rotationUnit:    rotationUnit,
		defaultSeverity: defaultSeverity,
		ctx:             ctx,
		cancel:          cancel,
		wg:              wg,
		current:         nil,
	}

	if err := rot.init(); err != nil {
		return nil, err
	}

	return rot, nil
}

func (l *LogRotator) Close() {
	l.cancel()
	l.wg.Wait()
}

// implementing io.Writer interface
func (l *LogRotator) Write(b []byte) (n int, err error) {
	if l.current == nil {
		return 0, errors.New("current logger is nil")
	}

	return l.writeBytes(l.current.UnderlyingLogger, l.defaultSeverity, b)
}

func (l *LogRotator) Log(sev severity.Severity, v ...any) error {
	if l.current == nil {
		return errors.New("current logger is nil")
	}

	return l.write(l.current.UnderlyingLogger, sev, v...)
}

func (l *LogRotator) init() error {
	if err := l.rotate(); err != nil {
		return err
	}

	return nil
}

func (l *LogRotator) fin() {
	if l.current != nil {
		if err := l.current.F.Close(); err != nil {
			// TODO: do nothing.. default log?
		}

		l.current = nil
	}
}

func (l *LogRotator) rotate() error {
	var buf bytes.Buffer

	now := l.now()

	if err := l.filePathTmpl.Execute(
		&buf,
		tmplData{
			TimeData: now,
			AppName:  l.appName,
			Addr:     l.addr,
		}); err != nil {

		return err
	}

	if l.current != nil && l.current.FileName == buf.String() {
		// no need to rotate
		return nil
	}

	newFile, err := l.openFile(buf.String())
	if err != nil {
		return err
	}

	prev = l.current
	l.current = &logger{
		FileName:         buf.String(),
		F:                newFile,
		UnderlyingLogger: log.New(newFile, "", 0),
	}

	l.write(l.current.UnderlyingLogger, l.defaultSeverity, "Log Opened.")

	if prev != nil {
		if err := prev.F.Close(); err != nil {
			// TODO: do nothing.. default log?
		}
	}

	return nil
}

func (l *LogRotator) now() timeData {
	t := time.Now().Round(0)

	if t.Location() != time.UTC {
		var base time.Time

		base = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
		base = base.Truncate(time.Duration(l.rotationUnit))
		base = time.Date(base.Year(), base.Month(), base.Day(), base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), t.Location()).Round(0)
		t = base
	}

	return timeData{
		Year:      t.Year(),
		YearShort: t.Year() % 100,
		Month:     int(t.Month()),
		Day:       t.Day(),
		Hour:      t.Hour(),
		Minute:    t.Minute(),
		Second:    t.Second(),
	}
}

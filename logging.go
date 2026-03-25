package fixzero

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Log interface {
	OnIncoming(string)
	OnOutgoing(string)
	OnEvent(string)
	OnError(string)
}

type NullLog struct{}

func (n *NullLog) OnIncoming(msg string) {}
func (n *NullLog) OnOutgoing(msg string) {}
func (n *NullLog) OnEvent(event string)  {}
func (n *NullLog) OnError(err string)    {}

var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[36m"
)

type ScreenLog struct {
	useColors bool
	mu        sync.Mutex
}

func NewScreenLog(useColors bool) *ScreenLog {
	return &ScreenLog{useColors: useColors}
}

func (s *ScreenLog) print(prefix, msg string, color string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if s.useColors {
		fmt.Fprintf(os.Stdout, "%s [%s] %s%s%s\n", timestamp, prefix, color, msg, colorReset)
	} else {
		fmt.Fprintf(os.Stdout, "%s [%s] %s\n", timestamp, prefix, msg)
	}
}

func (s *ScreenLog) OnIncoming(msg string) {
	s.print("INCOMING", msg, colorBlue)
}

func (s *ScreenLog) OnOutgoing(msg string) {
	s.print("OUTGOING", msg, colorGreen)
}

func (s *ScreenLog) OnEvent(event string) {
	s.print("EVENT", event, colorYellow)
}

func (s *ScreenLog) OnError(err string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if s.useColors {
		fmt.Fprintf(os.Stderr, "%s [ERROR] %s%s%s\n", timestamp, colorRed, err, colorReset)
	} else {
		fmt.Fprintf(os.Stderr, "%s [ERROR] %s\n", timestamp, err)
	}
}

type FileLog struct {
	incomingFile *os.File
	outgoingFile *os.File
	eventsFile   *os.File
	errorsFile   *os.File

	incomingWriter io.Writer
	outgoingWriter io.Writer
	eventsWriter   io.Writer
	errorsWriter   io.Writer

	mu        sync.Mutex
	dir       string
	maxSize   int64
	useColors bool
}

type FileLogOption func(*FileLog)

func WithLogDir(dir string) FileLogOption {
	return func(f *FileLog) { f.dir = dir }
}

func WithMaxSize(maxSize int64) FileLogOption {
	return func(f *FileLog) { f.maxSize = maxSize }
}

func WithFileColors(useColors bool) FileLogOption {
	return func(f *FileLog) { f.useColors = useColors }
}

func NewFileLog(opts ...FileLogOption) (*FileLog, error) {
	fl := &FileLog{
		dir:     ".",
		maxSize: 10 * 1024 * 1024,
	}

	for _, opt := range opts {
		opt(fl)
	}

	if err := os.MkdirAll(fl.dir, 0755); err != nil {
		return nil, err
	}

	incoming, err := os.OpenFile(filepath.Join(fl.dir, "incoming.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fl.incomingFile = incoming
	fl.incomingWriter = incoming

	outgoing, err := os.OpenFile(filepath.Join(fl.dir, "outgoing.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fl.Close()
		return nil, err
	}
	fl.outgoingFile = outgoing
	fl.outgoingWriter = outgoing

	events, err := os.OpenFile(filepath.Join(fl.dir, "events.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fl.Close()
		return nil, err
	}
	fl.eventsFile = events
	fl.eventsWriter = events

	errors, err := os.OpenFile(filepath.Join(fl.dir, "errors.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fl.Close()
		return nil, err
	}
	fl.errorsFile = errors
	fl.errorsWriter = errors

	return fl, nil
}

func (f *FileLog) rotateFile(file *os.File, path string) error {
	file.Close()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	newPath := fmt.Sprintf("%s.%s", path, timestamp)

	if err := os.Rename(path, newPath); err != nil {
		return err
	}

	newFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	*file = *newFile
	return nil
}

func (f *FileLog) writeWithRotate(writer *os.File, path, msg string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("%s %s\n", timestamp, msg)

	stat, err := writer.Stat()
	if err == nil && f.maxSize > 0 && stat.Size() > f.maxSize {
		f.rotateFile(writer, path)
	}

	writer.WriteString(entry)
}

func (f *FileLog) OnIncoming(msg string) {
	f.writeWithRotate(f.incomingFile, filepath.Join(f.dir, "incoming.log"), msg)
}

func (f *FileLog) OnOutgoing(msg string) {
	f.writeWithRotate(f.outgoingFile, filepath.Join(f.dir, "outgoing.log"), msg)
}

func (f *FileLog) OnEvent(event string) {
	f.writeWithRotate(f.eventsFile, filepath.Join(f.dir, "events.log"), event)
}

func (f *FileLog) OnError(err string) {
	f.writeWithRotate(f.errorsFile, filepath.Join(f.dir, "errors.log"), err)
}

func (f *FileLog) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errs []error

	if f.incomingFile != nil {
		if err := f.incomingFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if f.outgoingFile != nil {
		if err := f.outgoingFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if f.eventsFile != nil {
		if err := f.eventsFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if f.errorsFile != nil {
		if err := f.errorsFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing log files: %v", errs)
	}
	return nil
}

var stringPool = sync.Pool{
	New: func() interface{} {
		s := make([]byte, 0, 256)
		return &s
	},
}

func GetString() []byte {
	bp := stringPool.Get().(*[]byte)
	*bp = (*bp)[:0]
	return *bp
}

func PutString(s []byte) {
	if cap(s) > 0 && cap(s) <= 4096 {
		stringPool.Put(&s)
	}
}

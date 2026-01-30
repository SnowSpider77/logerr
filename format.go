package logerr

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type TimePosition uint8

const (
	TimeOff TimePosition = iota
	TimeBefore
	TimeAfter
)

type PathMode uint8

const (
	PathAbsolute   PathMode = iota
	PathWorkingDir          // relative to os.Getwd()
	PathExecutable          // relative to os.Executable()
)

type Config struct {
	TimeFormat   string
	TimePosition TimePosition
	Location     *time.Location

	LogPath  string
	PathMode PathMode
	Append   bool
}

func resolveLogPath(cfg Config) (string, error) {
	switch cfg.PathMode {
	case PathAbsolute:
		return cfg.LogPath, nil

	case PathWorkingDir:
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, cfg.LogPath), nil

	case PathExecutable:
		exe, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.Join(filepath.Dir(exe), cfg.LogPath), nil

	default:
		return "", fmt.Errorf("unknown PathMode")
	}
}

func openLogFile(path string, append bool) (*os.File, error) {
	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(path, flags, 0644)
}

var (
	writer      atomic.Value // stores io.Writer
	configValue atomic.Value // stores Config
)

func init() {
	writer.Store(io.Discard)
	configValue.Store(Config{
		TimeFormat:   time.RFC3339,
		TimePosition: TimeBefore,
		Location:     time.Local,
		LogPath:      "",
		PathMode:     PathAbsolute,
		Append:       true,
	})
}

func SetConfig(c Config) error {
	// Normalize time defaults first
	if c.TimeFormat == "" {
		c.TimeFormat = time.RFC3339
	}
	if c.Location == nil {
		c.Location = time.Local
	}

	// If empty path => disable file logging
	if c.LogPath == "" {
		swapWriter(io.Discard) // closes previous file if needed
		configValue.Store(c)
		return nil
	}

	path, err := resolveLogPath(c)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := openLogFile(path, c.Append)
	if err != nil {
		return err
	}

	// Swap writer and close old file if it was a file
	swapWriter(f)

	configValue.Store(c)
	return nil
}

func swapWriter(w io.Writer) {
	old := writer.Swap(w) // atomic swap
	if oldFile, ok := old.(*os.File); ok && oldFile != nil {
		_ = oldFile.Close()
	}
}

func GetConfig() Config {
	return configValue.Load().(Config)
}

package appconfig

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultListenAddress    = "127.0.0.1:7331"
	defaultDataDirectory    = "GuidedStudy"
	defaultDatabaseFilename = "guided-study.db"
	defaultConverterName    = "pdf-converter"
	defaultHeadless         = false

	windowsGOOS      = "windows"
	windowsExeSuffix = ".exe"
)

type Config struct {
	Listen        string
	DatabasePath  string
	ConverterPath string
	Headless      bool
}

// Parse reads command-line configuration and applies defaults.
func Parse(args []string) (Config, error) {
	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, err
	}

	set := flag.NewFlagSet("noggin-mcp", flag.ContinueOnError)

	// Flags override the default values in cfg.
	set.StringVar(&cfg.Listen, "listen", cfg.Listen, "HTTP listen address")
	set.StringVar(&cfg.DatabasePath, "database", cfg.DatabasePath, "SQLite database path")
	set.StringVar(&cfg.ConverterPath, "converter", cfg.ConverterPath, "PDF converter executable")
	set.BoolVar(
		&cfg.Headless,
		"headless",
		cfg.Headless,
		"run the MCP service without the desktop tray",
	)

	if err := set.Parse(args); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func defaultConfig() (Config, error) {
	// Store application data in the current user's data directory.
	base, err := userDataDirectory()
	if err != nil {
		return Config{}, err
	}

	// Find the PDF converter beside the desktop executable.
	exe, err := os.Executable()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Listen:        defaultListenAddress,
		DatabasePath:  filepath.Join(base, defaultDataDirectory, defaultDatabaseFilename),
		ConverterPath: filepath.Join(filepath.Dir(exe), converterFilename()),
		Headless:      defaultHeadless,
	}, nil
}

// userDataDirectory returns the per-user data root.
func userDataDirectory() (string, error) {
	// Windows keeps application data out of the roaming profile.
	if runtime.GOOS == windowsGOOS {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
		return local, nil
	}

	// macOS uses Application Support, Linux uses XDG config home.
	return os.UserConfigDir()
}

// converterFilename returns the bundled converter name.
func converterFilename() string {
	if runtime.GOOS == windowsGOOS {
		return defaultConverterName + windowsExeSuffix
	}
	return defaultConverterName
}

// Validate checks required values and paths.
func (c Config) Validate() error {
	_, _, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	if strings.TrimSpace(c.DatabasePath) == "" || strings.TrimSpace(c.ConverterPath) == "" {
		return errors.New("database and converter paths must not be blank")
	}

	info, err := os.Stat(c.ConverterPath)
	if err != nil {
		return fmt.Errorf("PDF converter is not accessible at %q: %w", c.ConverterPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("PDF converter path %q is a directory", c.ConverterPath)
	}

	return nil
}

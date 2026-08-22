package appconfig

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultListenAddress     = "127.0.0.1:7331"
	defaultDataDirectory     = "GuidedStudy"
	defaultDatabaseFilename  = "guided-study.db"
	defaultConverterFilename = "pdf-converter.exe"
	defaultHeadless          = false
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

	set := flag.NewFlagSet("guided-study", flag.ContinueOnError)

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
	// Store application data in the current user's local app-data directory.
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return Config{}, errors.New("LOCALAPPDATA is not set")
	}

	// Find the PDF converter beside the desktop executable.
	exe, err := os.Executable()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Listen:        defaultListenAddress,
		DatabasePath:  filepath.Join(local, defaultDataDirectory, defaultDatabaseFilename),
		ConverterPath: filepath.Join(filepath.Dir(exe), defaultConverterFilename),
		Headless:      defaultHeadless,
	}, nil
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

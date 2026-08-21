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

type Config struct {
	Listen        string
	DatabasePath  string
	ConverterPath string
	Headless      bool
}

// Parse builds a configuration from platform-specific defaults and command-line
// overrides. The resulting configuration is validated before it is returned.
func Parse(args []string) (Config, error) {
	defaults, err := defaultPaths()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	set := flag.NewFlagSet("guided-study", flag.ContinueOnError)

	// Every flag writes directly into cfg. Defaults discovered above are used
	// only when the caller does not provide the corresponding flag.
	set.StringVar(&cfg.Listen, "listen", "127.0.0.1:7331", "HTTP listen address")
	set.StringVar(&cfg.DatabasePath, "database", defaults.database, "SQLite database path")
	set.StringVar(&cfg.ConverterPath, "converter", defaults.converter, "PDF converter executable")
	set.BoolVar(&cfg.Headless, "headless", false, "run the MCP service without the desktop tray")

	if err := set.Parse(args); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type paths struct{ database, converter string }

func defaultPaths() (paths, error) {
	// Application data belongs under the current Windows user's local app-data
	// directory. Without it, there is no safe default database location.
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return paths{}, errors.New("LOCALAPPDATA is not set")
	}

	// The frozen PDF converter is distributed beside the desktop executable.
	exe, err := os.Executable()
	if err != nil {
		return paths{}, err
	}

	converter := filepath.Join(filepath.Dir(exe), "pdf-converter.exe")

	return paths{
		database:  filepath.Join(local, "GuidedStudy", "guided-study.db"),
		converter: converter,
	}, nil
}

// Validate rejects unsafe network exposure and unusable required paths.
func (c Config) Validate() error {
	// This is a desktop-local service. Binding it to a non-loopback interface
	// would expose its HTTP/MCP endpoints to other machines on the network.
	host, _, err := net.SplitHostPort(c.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	if !isLoopback(host) {
		return fmt.Errorf("refusing non-loopback listen host %q", host)
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

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/andgate/guided-study-desktop-mcp/internal/appconfig"
	"github.com/andgate/guided-study-desktop-mcp/internal/importer"
	"github.com/andgate/guided-study-desktop-mcp/internal/localserver"
	"github.com/andgate/guided-study-desktop-mcp/internal/mcpserver"
	"github.com/andgate/guided-study-desktop-mcp/internal/store"
)

//go:embed assets/tray-icon.png
var trayIconBytes []byte

var trayIconResource = fyne.NewStaticResource("guided-study-tray.png", trayIconBytes)

// main translates a startup failure into a logged message and a nonzero exit status.
func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("guided-study stopped", "error", err)
		os.Exit(1)
	}
}

// run assembles the application and owns its process-level lifecycle.
func run(args []string) error {
	// Load and validate settings before creating any application resources.
	cfg, err := appconfig.Parse(args)
	if err != nil {
		return err
	}

	// Prepare the data directory before opening the canonical SQLite store.
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	// Tie application lifetime to Ctrl+C and ordinary process termination.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Open durable state before constructing services that depend on it.
	st, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer st.Close()

	// Assemble the importer, MCP service, and local HTTP transport.
	imp := importer.New(importer.Config{
		ConverterPath: cfg.ConverterPath,
		DPI:           200,
		JPEGQuality:   90,
	})
	mcpService := mcpserver.New(st, imp, logger)
	httpServer := localserver.New(cfg.Listen, mcpService.Handler(logger))
	serverErrors := startHTTPServer(httpServer, logger, cfg.DatabasePath)

	// Hand lifecycle control to either the terminal or desktop interface.
	if cfg.Headless {
		return waitHeadless(ctx, httpServer, serverErrors)
	}
	return runDesktop(ctx, cancel, cfg, httpServer, serverErrors)
}

// startHTTPServer reports unexpected listener failures. A normal shutdown
// closes the returned channel without sending an error.
func startHTTPServer(server *http.Server, logger *slog.Logger, databasePath string) <-chan error {
	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"MCP server listening",
			"url", "http://"+server.Addr+"/mcp",
			"database", databasePath,
		)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	return serverErrors
}

// waitHeadless blocks until the listener fails or the process receives a stop signal.
func waitHeadless(ctx context.Context, server *http.Server, serverErrors <-chan error) error {
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		return shutdownServer(server)
	}
}

// runDesktop owns the Fyne event loop and coordinates it with the HTTP server.
func runDesktop(ctx context.Context, cancel context.CancelFunc, cfg appconfig.Config, server *http.Server, serverErrors <-chan error) (resultErr error) {
	// Once desktop startup begins, every return path owns graceful server cleanup.
	defer func() {
		resultErr = errors.Join(resultErr, shutdownServer(server))
	}()

	// Construct the visible status window before exposing tray actions.
	a := app.NewWithID("com.andgate.guided-study")
	a.SetIcon(trayIconResource)
	window, status, endpoint := newStatusWindow(a, cfg)

	// Give the tray's Quit action one path for cancellation and graceful shutdown.
	quit := func() {
		status.SetText("Stopping")
		cancel()
		a.Quit()
	}
	if err := configureSystemTray(a, window, endpoint, quit); err != nil {
		return err
	}

	// Watch the server while Fyne owns the current thread.
	desktopResult := monitorDesktop(ctx, a, window, status, serverErrors)

	window.Show()
	a.Run()

	// Complete process cleanup after the UI event loop exits.
	cancel()

	select {
	case err := <-desktopResult:
		if err != nil {
			return err
		}
	default:
	}
	return nil
}

// newStatusWindow builds the read-only service status interface.
func newStatusWindow(a fyne.App, cfg appconfig.Config) (fyne.Window, *widget.Label, string) {
	window := a.NewWindow("Guided Study")
	endpoint := "http://" + cfg.Listen + "/mcp"

	// Build the status and connection fields displayed to the user.
	status := widget.NewLabel("Running")
	status.TextStyle = fyne.TextStyle{Bold: true}

	endpointEntry := widget.NewEntry()
	endpointEntry.SetText(endpoint)
	endpointEntry.Disable()

	databaseEntry := widget.NewEntry()
	databaseEntry.SetText(cfg.DatabasePath)
	databaseEntry.Disable()

	// Compose the controls into the final vertical window layout.
	copyButton := widget.NewButton("Copy endpoint", func() {
		a.Clipboard().SetContent(endpoint)
	})
	connectionForm := widget.NewForm(
		widget.NewFormItem("Endpoint", endpointEntry),
		widget.NewFormItem("Database", databaseEntry),
	)
	content := container.NewVBox(
		widget.NewLabel("Guided Study MCP service"),
		status,
		connectionForm,
		copyButton,
		widget.NewLabel("Closing this window keeps the service running. Use Quit from the tray to stop it."),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(620, 260))
	window.SetCloseIntercept(window.Hide)

	return window, status, endpoint
}

// configureSystemTray connects the persistent background service to desktop actions.
func configureSystemTray(a fyne.App, window fyne.Window, endpoint string, quit func()) error {
	desktopApp, ok := a.(desktop.App)
	if !ok {
		return errors.New("Fyne desktop system tray is unavailable")
	}

	menu := fyne.NewMenu(
		"Guided Study",
		fyne.NewMenuItem("Show status", window.Show),
		fyne.NewMenuItem("Copy MCP endpoint", func() {
			a.Clipboard().SetContent(endpoint)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", quit),
	)
	desktopApp.SetSystemTrayMenu(menu)
	desktopApp.SetSystemTrayIcon(trayIconResource)

	return nil
}

// monitorDesktop coordinates process cancellation and HTTP listener failures
// with Fyne's UI thread, then reports the reason the desktop loop stopped.
func monitorDesktop(
	ctx context.Context,
	a fyne.App,
	window fyne.Window,
	status *widget.Label,
	serverErrors <-chan error,
) <-chan error {
	result := make(chan error, 1)

	go func() {
		var resultErr error

		select {
		case err := <-serverErrors:
			if err != nil {
				resultErr = err
				fyne.Do(func() {
					status.SetText("Server error: " + err.Error())
					window.Show()
				})
			}
		case <-ctx.Done():
		}

		result <- resultErr
		close(result)
		fyne.Do(a.Quit)
	}()

	return result
}

// shutdownServer gives active HTTP requests a bounded period to finish.
func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}

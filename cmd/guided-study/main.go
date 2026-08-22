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

// main logs startup failures.
func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("guided-study stopped", "error", err)
		os.Exit(1)
	}
}

// run starts the application.
func run(args []string) error {
	// Load the application settings.
	cfg, err := appconfig.Parse(args)
	if err != nil {
		return err
	}

	// Create the database directory.
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	// Stop when the process receives a signal.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Open the database.
	st, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer st.Close()

	// Create the services and HTTP server.
	imp := importer.New(importer.Config{
		ConverterPath: cfg.ConverterPath,
	})
	mcpService := mcpserver.New(st, imp, logger)
	httpServer := localserver.New(cfg.Listen, mcpService.Handler(logger))
	serverErrors := startHTTPServer(httpServer, logger, cfg.DatabasePath)

	// Start the selected interface.
	if cfg.Headless {
		return waitHeadless(ctx, httpServer, serverErrors)
	}
	return runDesktop(ctx, cancel, cfg, httpServer, serverErrors)
}

// startHTTPServer runs the HTTP listener.
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

// waitHeadless waits for shutdown or failure.
func waitHeadless(ctx context.Context, server *http.Server, serverErrors <-chan error) error {
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		return shutdownServer(server)
	}
}

// runDesktop runs the desktop interface.
func runDesktop(
	ctx context.Context,
	cancel context.CancelFunc,
	cfg appconfig.Config,
	server *http.Server,
	serverErrors <-chan error,
) (resultErr error) {
	// Stop the HTTP server before returning.
	defer func() {
		resultErr = errors.Join(resultErr, shutdownServer(server))
	}()

	// Create the status window.
	a := app.NewWithID("com.andgate.guided-study")
	a.SetIcon(trayIconResource)
	window, status, endpoint := newStatusWindow(a, cfg)

	// Stop the application from the tray.
	quit := func() {
		status.SetText("Stopping")
		cancel()
		a.Quit()
	}
	if err := configureSystemTray(a, window, endpoint, quit); err != nil {
		return err
	}

	// Watch the server for failures.
	desktopResult := monitorDesktop(ctx, a, window, status, serverErrors)

	window.Show()
	a.Run()

	// Cancel remaining background work.
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

// newStatusWindow creates the status window.
func newStatusWindow(a fyne.App, cfg appconfig.Config) (fyne.Window, *widget.Label, string) {
	window := a.NewWindow("Guided Study")
	endpoint := "http://" + cfg.Listen + "/mcp"

	// Create the status fields.
	status := widget.NewLabel("Running")
	status.TextStyle = fyne.TextStyle{Bold: true}

	endpointEntry := widget.NewEntry()
	endpointEntry.SetText(endpoint)
	endpointEntry.Disable()

	databaseEntry := widget.NewEntry()
	databaseEntry.SetText(cfg.DatabasePath)
	databaseEntry.Disable()

	// Build the window layout.
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
		widget.NewLabel(
			"Closing this window keeps the service running. Use Quit from the tray to stop it.",
		),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(620, 260))
	window.SetCloseIntercept(window.Hide)

	return window, status, endpoint
}

// configureSystemTray creates the tray menu.
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

// monitorDesktop watches for shutdown and errors.
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

// shutdownServer waits for active requests.
func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}

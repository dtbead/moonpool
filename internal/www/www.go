package www

import (
	"context"
	"embed"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/dtbead/moonpool/api"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/labstack/echo/v4"
)

//go:embed assets
var folderAssets embed.FS

//go:embed templates
var folderTemplates embed.FS

type WWW struct {
	echo   *echo.Echo
	api    *api.API
	log    *slog.Logger
	config Config
}

type searchOptions struct {
	Query string
	Sort  string
	Order string
}

type Config struct {
	DynamicWebReloading     bool
	DynamicWebReloadingPath string
	LogLevel                slog.Level
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func New(apiConfig api.Config, webConfig Config) (WWW, error) {
	logMain := log.New(webConfig.LogLevel)
	logAPI := logMain.WithGroup("api")

	api, err := api.Open(apiConfig, logAPI)
	if err != nil {
		return WWW{}, err
	}

	w := WWW{
		config: webConfig,
		echo:   echo.New(),
		log:    logMain,
		api:    api,
	}

	return w, nil
}

func (w WWW) Start(ListenAddress string) error {
	w.init()

	if w.config.DynamicWebReloading {
		w.echo.Static("/", w.config.DynamicWebReloadingPath)
	} else {
		w.echo.StaticFS("/", folderAssets)

		t, err := template.ParseFS(folderTemplates, "templates/*")
		if err != nil {
			return err
		}
		w.echo.Renderer = &Template{t}
	}

	return w.echo.Start(ListenAddress)
}

func (w WWW) init() {
	w.echo.Static("media", w.api.Config.MediaLocation)
	w.echo.HideBanner = true
	w.echo.HTTPErrorHandler = w.errorHandler

	w.Thumbnail()
	w.Post()
	w.Browse()
}

func (w WWW) errorHandler(err error, c echo.Context) {
	log := w.log.With(slog.Group("web_ui",
		slog.Any("error", err),
		slog.Any("time", time.Now()),
		slog.String("ip", c.RealIP()),
		slog.String("url", c.Request().RequestURI),
		slog.String("method", c.Request().Method),
		slog.String("user-agent", c.Request().UserAgent()),
		slog.Any("form", c.Request().Form),
	))

	log.Error("error", slog.Any("error", err))

	he, ok := err.(*echo.HTTPError)
	if !ok {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	if !c.Response().Committed {
		err = c.String(he.Code, he.Message.(string))
		if err != nil {
			log.Error("echo_error", slog.Any("error", err))
		}
	}
}

func parseSearchOptions(c echo.Context) searchOptions {
	searchOptions := searchOptions{
		Sort:  c.FormValue("sort"),
		Query: c.FormValue("query"),
		Order: c.FormValue("order"),
	}

	if searchOptions.Sort == "" {
		searchOptions.Sort = "imported"
	}

	return searchOptions
}

func (w WWW) Shutdown(ctx context.Context) error {
	return w.echo.Shutdown(ctx)
}

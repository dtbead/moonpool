package www

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"text/template"

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

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Config struct {
	DynamicWebReloading     bool
	DynamicWebReloadingPath string
	LogLevel                slog.Level
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
	w.echo.HTTPErrorHandler = customHTTPErrorHandler

	w.Thumbnail()
	w.Post()
	w.Browse()
}

func (w WWW) Shutdown(ctx context.Context) error {
	return w.echo.Shutdown(ctx)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	c.Logger().Error(err)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	errorPage := fmt.Sprintf("templates/%d.html", code) // TODO: add other *.html error code support
	if err := c.File(errorPage); err != nil {
		if !errors.Is(err, echo.ErrNotFound) {
			return
		} else {
			c.Logger().Error(err)
		}
	}
}

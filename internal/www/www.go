package www

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

//go:embed assets
var folderAssets embed.FS

//go:embed templates
var folderTemplates embed.FS

type WWW struct {
	echo   *echo.Echo
	api    *api.API
	config Config
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Config struct {
	DynamicWebReloading bool
}

func New(a *api.API, c Config) *WWW {
	w := WWW{
		echo:   echo.New(),
		api:    a,
		config: c,
	}

	w.echo.Static("media", w.api.Config.MediaLocation)
	w.echo.HideBanner = true
	w.echo.HTTPErrorHandler = customHTTPErrorHandler

	w.init()
	return &w
}

func (w WWW) Start(ListenAddress string) error {
	if w.config.DynamicWebReloading {
		w.echo.Static("/", "assets")
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
	w.Thumbnail()
	w.Post()
	w.Browse()
}

func (w WWW) Shutdown() error {
	return w.echo.Shutdown(context.Background())
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

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
	e *echo.Echo
	a *api.API
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func New(a *api.API) *WWW {
	w := WWW{
		e: echo.New(),
		a: a,
	}

	w.e.HideBanner = true
	w.e.HTTPErrorHandler = customHTTPErrorHandler

	return &w
}

func (w WWW) Start(ListenAddress string) error {
	w.e.StaticFS("/", folderAssets)
	w.e.Static("media", w.a.Config.MediaLocation)

	t, err := template.ParseFS(folderTemplates, "templates/*")
	if err != nil {
		return err
	}

	w.e.Renderer = &Template{t}
	w.Post()
	w.Browse()

	return w.e.Start(ListenAddress)
}

func (w WWW) Shutdown() error {
	return w.e.Shutdown(context.Background())
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	errorPage := fmt.Sprintf("templates/%d.html", code) // TODO: add other *.html error code support
	if err := c.File(errorPage); err != nil {
		if !errors.Is(err, echo.ErrNotFound) {
			c.Logger().Error(err)
		}
	}
}

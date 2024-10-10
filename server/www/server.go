package www

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

var rootDir string

type WWW struct {
	e *echo.Echo
	a *api.API
}

func New(a *api.API) *WWW {
	w := WWW{
		e: echo.New(),
		a: a,
	}
	w.init()
	return &w
}

func (w WWW) Start(ListenAddress string) error {
	return w.e.Start(ListenAddress)
}

func (w WWW) Shutdown() error {
	return w.e.Shutdown(context.Background())
}

func (w WWW) init() {
	workingDirectory, _ := os.Getwd()

	w.e.HideBanner = true
	w.e.Static("/", workingDirectory+"/assets")
	w.e.Static("media", w.a.Config.MediaLocation)

	w.e.HTTPErrorHandler = customHTTPErrorHandler

	w.Post()
	w.Browse()
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(err)
	errorPage := fmt.Sprintf("%s/templates/%d.html", rootDir, code) // TODO: add other *.html error code support
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

package www

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

var rootDir string

type WWW struct {
	E *echo.Echo
	A *api.API
}

func New(a api.API, mediaPath string) WWW {
	WW := WWW{
		E: echo.New(),
		A: &a,
	}

	WW.Init(mediaPath)
	return WW
}

func (w WWW) Init(mediaPath string) {
	rootDir = projectDirectory()
	w.E.Static("/", rootDir+"/assets")
	w.E.Static("media", mediaPath)

	w.E.HTTPErrorHandler = customHTTPErrorHandler

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

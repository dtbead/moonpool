package www

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

var rootDir string
var error404 string

type WWW struct {
	E *echo.Echo
	A *api.API
}

func New(a api.API) WWW {
	WW := WWW{
		E: echo.New(),
		A: &a,
	}

	WW.Init()
	return WW
}

func (w WWW) Init() {
	rootDir = projectDirectory()
	w.E.Static("/static", "assets")

	f, err := os.ReadFile(rootDir + "\\templates\\404.html")
	if err != nil {
		fmt.Println("failed to initialize WWW server!", err)
		os.Exit(1)
	}
	error404 = string(f)

	w.Post()
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func projectDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	return basepath
}

func (w WWW) Serve404(c echo.Context) {
	c.HTML(http.StatusNotFound, string(error404))
}

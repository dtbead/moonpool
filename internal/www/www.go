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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed web/**
var webFolder embed.FS

type WWW struct {
	echo    *echo.Echo
	api     *api.API
	logMain *slog.Logger
	logAPI  *slog.Logger
	config  Config
}

type searchOptions struct {
	Query                  string
	Sort                   string
	Order                  string
	PageAmount, PageOffset int64
}

type Config struct {
	DynamicWebReloading     bool
	DynamicWebReloadingPath string
	Log                     *slog.Logger
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func New(a *api.API, webConfig Config) (WWW, error) {
	logMain := webConfig.Log
	logAPI := webConfig.Log.WithGroup("api")

	w := WWW{
		config:  webConfig,
		echo:    echo.New(),
		logMain: logMain,
		logAPI:  logAPI,
		api:     a,
	}

	return w, nil
}

func (w WWW) Start(ListenAddress string) error {
	w.init()

	if w.config.DynamicWebReloading {
		w.echo.Static("/", w.config.DynamicWebReloadingPath)
	} else {
		w.echo.StaticFS("/", webFolder)

		t := template.New("").Funcs(templateFuncMap)
		t, err := t.ParseFS(webFolder, "web/templates/*")
		if err != nil {
			return err
		}

		t, err = t.ParseFS(webFolder, "web/assets/**")
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
	w.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://127.0.0.1:9995"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	w.Root()
	w.Post()
	w.Browse()
	w.Thumbnail()
}

func (w WWW) Root() {
	w.echo.Pre(middleware.Rewrite(map[string]string{
		"/": "/browse",
	}))
}

func (w WWW) errorHandler(err error, c echo.Context) {
	log := w.logMain.With(
		slog.Any("error", err),
		slog.Any("time", time.Now()),
		slog.String("ip", c.RealIP()),
		slog.String("url", c.Request().RequestURI),
		slog.String("method", c.Request().Method),
		slog.String("user-agent", c.Request().UserAgent()),
		slog.Any("form", c.Request().Form),
	)

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

var templateFuncMap = map[string]any{
	"add": add,
}

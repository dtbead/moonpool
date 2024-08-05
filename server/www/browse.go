package www

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
)

func (w WWW) Browse() {
	w.E.GET("browse", func(c echo.Context) error {
		tmpl := &Template{
			templates: template.Must(template.ParseFiles(projectDirectory() + "/templates/browse.html")),
		}
		w.E.Renderer = tmpl

		if err := c.Render(http.StatusOK, "browse.html", nil); err != nil {
			fmt.Printf("error rendering Post. %v\n", err)
			return err
		}
		return nil
	})
}

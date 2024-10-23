package www

import (
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
)

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {
		if w.config.DynamicWebReloading {
			tmpl := &Template{
				templates: template.Must(template.ParseFiles(getProjectDirectory() + "/templates/browse.html")),
			}
			w.echo.Renderer = tmpl
		}

		page, err := w.api.GetPage(context.Background(), 50, 0)
		if err != nil {
			return err
		}

		archive_ids := make([]int64, len(page))
		for i, v := range page {
			archive_ids[i] = v.ArchiveID
		}

		if err := c.Render(http.StatusOK, "browse.html", map[string]interface{}{
			"entries": archive_ids,
		}); err != nil {
			fmt.Printf("error rendering browse.html. %v\n", err)
			return err
		}
		return nil
	})
}

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
		ctx := context.Background()
		if w.config.DynamicWebReloading {
			tmpl := &Template{
				templates: template.Must(template.ParseFiles(getProjectDirectory() + "/templates/browse.html")),
			}
			w.echo.Renderer = tmpl
		}

		page, err := w.api.GetPage(ctx, "imported", 40, 0)
		if err != nil {
			return err
		}

		pageTags, err := w.api.GetTagsByRange(ctx, 0, 10, 0)
		if err != nil {
			return err
		}

		archive_ids := make([]int64, len(page))
		for i, v := range page {
			archive_ids[i] = v.ID
		}

		if err := c.Render(http.StatusOK, "browse.html", map[string]interface{}{
			"entries": archive_ids,
			"tagList": pageTags,
		}); err != nil {
			fmt.Printf("error rendering browse.html. %v\n", err)
			return err
		}
		return nil
	})
}

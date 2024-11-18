package www

import (
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

const DEFAULT_PAGES_MAX int = 50

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {

		ctx := context.Background()
		if w.config.DynamicWebReloading {
			tmpl := &Template{
				templates: template.Must(template.ParseFiles(getProjectDirectory() + "/templates/browse.html")),
			}
			w.echo.Renderer = tmpl
		}

		searchSort := c.FormValue("sort")
		searchQuery := c.FormValue("query")

		if searchSort == "" {
			searchSort = "imported"
		}

		if searchQuery != "" {
			res, err := w.api.QueryTags(ctx, searchSort, api.BuildQuery(searchQuery))
			if err != nil {
				return err
			}

			tags, err := w.api.GetTagsByList(ctx, res)
			if err != nil {
				return err
			}

			if err := c.Render(http.StatusOK, "browse.html", map[string]interface{}{
				"entries": res,
				"tagList": tags,
				"query":   "",
			}); err != nil {
				return err
			}
			return nil
		}

		pageAmount := int(stringToInt64(c.FormValue("amount")))
		if pageAmount <= 0 || pageAmount > DEFAULT_PAGES_MAX {
			pageAmount = DEFAULT_PAGES_MAX
		}

		pageOffset := int(stringToInt64(c.FormValue("offset")))
		if pageOffset < 0 {
			pageOffset = 0
		}

		page, err := w.api.GetPage(ctx, "imported", pageAmount, pageOffset)
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
			"query":   "",
		}); err != nil {
			fmt.Printf("error rendering browse.html. %v\n", err)
			return err
		}
		return nil
	})
}

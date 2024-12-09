package www

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

const DEFAULT_PAGES_MAX int64 = 50

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {
		if w.config.DynamicWebReloading {
			tmp, err := template.New("browse.html").Funcs(templateFuncMap).ParseFiles(
				w.config.DynamicWebReloadingPath + "/templates/browse.html")
			if err != nil {
				return err
			}

			w.echo.Renderer = &Template{tmp}
		}

		ctx := context.Background()
		searchOptions := parseSearchOptions(c)

		descedingOrder := true
		if strings.EqualFold(searchOptions.Order, "ascending") {
			descedingOrder = false
		}

		if searchOptions.Query != "" {
			res, err := w.api.QueryTags(ctx, strings.ToLower(searchOptions.Sort), strings.ToLower(searchOptions.Order), api.BuildQuery(searchOptions.Query))
			if err != nil {
				return err
			}

			tags, err := w.api.GetTagsByList(ctx, res)
			if err != nil {
				return err
			}

			if err := c.Render(http.StatusOK, "browse.html", map[string]interface{}{
				"entries":       res,
				"tagList":       tags,
				"searchOptions": searchOptions,
			}); err != nil {
				return err
			}
			return nil
		}

		searchOptions.PageAmount = stringToInt64(c.FormValue("amount"))
		if searchOptions.PageAmount <= 0 || searchOptions.PageAmount > DEFAULT_PAGES_MAX {
			searchOptions.PageAmount = DEFAULT_PAGES_MAX
		}

		searchOptions.PageOffset = stringToInt64(c.FormValue("offset"))
		if searchOptions.PageOffset < 0 {
			searchOptions.PageOffset = 0
		}

		page, err := w.api.GetPage(ctx, searchOptions.Sort, searchOptions.PageAmount, searchOptions.PageOffset, descedingOrder)
		if err != nil {
			return err
		}

		archive_ids := make([]int64, len(page))
		for i, v := range page {
			archive_ids[i] = v.ID
		}

		pageTags, err := w.api.GetTagsByList(ctx, archive_ids)
		if err != nil {
			return err
		}

		if err := c.Render(http.StatusOK, "browse.html", map[string]interface{}{
			"entries":       archive_ids,
			"tagList":       pageTags,
			"searchOptions": searchOptions,
		}); err != nil {
			fmt.Printf("error rendering browse.html. %v\n", err)
			return err
		}
		return nil
	})
}

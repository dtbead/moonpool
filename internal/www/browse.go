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

const DEFAULT_PAGES_MAX int = 50

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {
		ctx := context.Background()

		if w.config.DynamicWebReloading {
			tmp, err := template.ParseFiles(w.config.DynamicWebReloadingPath + "/templates/browse.html")
			if err != nil {
				return err
			}
			w.echo.Renderer = &Template{tmp}
		}

		searchOptions := parseSearchOptions(c)

		// TODO: fis desc button not persisting, then commit changes
		descedingOrder := true
		if strings.EqualFold(searchOptions.Order, "ascending") {
			descedingOrder = false
		}

		if searchOptions.Query != "" {
			res, err := w.api.QueryTags(ctx, searchOptions.Sort, api.BuildQuery(searchOptions.Query))
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

		pageAmount := int(stringToInt64(c.FormValue("amount")))
		if pageAmount <= 0 || pageAmount > DEFAULT_PAGES_MAX {
			pageAmount = DEFAULT_PAGES_MAX
		}

		pageOffset := int(stringToInt64(c.FormValue("offset")))
		if pageOffset < 0 {
			pageOffset = 0
		}

		page, err := w.api.GetPage(ctx, searchOptions.Sort, pageAmount, pageOffset, descedingOrder)
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

package www

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

const DEFAULT_PAGES_MAX int = 50

type searchOptions struct {
	Query string
	Sort  string
}

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if w.config.DynamicWebReloading {
			tmp, err := template.ParseFiles(w.config.DynamicWebReloadingPath + "/templates/browse.html")
			if err != nil {
				return err
			}
			w.echo.Renderer = &Template{tmp}
		}

		searchOptions := searchOptions{
			Sort:  c.FormValue("sort"),
			Query: c.FormValue("query"),
		}

		if searchOptions.Sort == "" {
			searchOptions.Sort = "imported"
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

		page, err := w.api.GetPage(ctx, searchOptions.Sort, pageAmount, pageOffset)
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

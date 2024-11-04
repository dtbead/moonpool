package www

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/internal/file"
	"github.com/labstack/echo/v4"
)

func (w WWW) Post() {
	w.echo.GET("post/entry/:id", func(c echo.Context) error {
		if w.config.DynamicWebReloading {
			tmpl := &Template{
				templates: template.Must(template.ParseFiles(getProjectDirectory() + "/templates/entry.html")),
			}
			w.echo.Renderer = tmpl
		}

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			return fmt.Errorf("invalid archive ID")
		}

		media, err := w.api.GetPath(context.TODO(), archive_id)
		if err != nil {
			return err
		}

		hashes, err := w.api.GetHashes(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		timestamps, err := w.api.GetTimestamps(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			missingTimestamps := 0
			if timestamps.DateCreated.IsZero() {
				missingTimestamps++
			}

			if timestamps.DateImported.IsZero() {
				missingTimestamps++
			}

			if timestamps.DateModified.IsZero() {
				missingTimestamps++
			}

			if missingTimestamps >= 3 {
				return err
			}

			fmt.Println("found parital timestamps")
		}

		tags, err := w.api.GetTags(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if err := c.Render(http.StatusOK, "entry.html", map[string]interface{}{
			"tags": tags,
			"hashes": map[string]string{
				"MD5":    file.ByteToHexString(hashes.MD5),
				"SHA1":   file.ByteToHexString(hashes.SHA1),
				"SHA256": file.ByteToHexString(hashes.SHA256),
			},
			"timestamps": map[string]string{
				"imported": timeToString(timestamps.DateImported.Local()),
				"modified": timeToString(timestamps.DateModified.Local()),
				"created":  timeToString(timestamps.DateCreated.Local()),
			},
			"media": media.FileRelative,
		}); err != nil {
			fmt.Printf("error rendering post. %v\n", err)
			return err
		}
		return nil
	})
}

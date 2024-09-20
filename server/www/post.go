package www

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/internal/file"
	"github.com/dtbead/moonpool/server"
	"github.com/labstack/echo/v4"
)

func (w WWW) Post() {
	w.E.GET("post/entry/:id", func(c echo.Context) error {
		tmpl := &Template{
			templates: template.Must(template.ParseFiles(projectDirectory() + "/templates/entry.html")),
		}
		w.E.Renderer = tmpl

		archive_id := server.ValidateArchiveID(*w.A, c.Param("id"))
		if archive_id == -1 {
			return server.ErrInvalidArchiveID
		}

		media, err := w.A.GetPath(context.TODO(), archive_id)
		if err != nil {
			return err
		}

		hashes, err := w.A.GetHashes(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		timestamps, err := w.A.GetTimestamps(context.TODO(), archive_id)
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

		tags, err := w.A.GetTags(context.TODO(), archive_id)
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
				"imported": timeToString(timestamps.DateImported),
				"modified": timeToString(timestamps.DateModified),
				"created":  timeToString(timestamps.DateCreated),
			},
			"media": media.FileRelative,
		}); err != nil {
			fmt.Printf("error rendering Post. %v\n", err)
			return err
		}
		return nil
	})
}

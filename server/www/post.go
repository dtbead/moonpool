package www

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/server"
	"github.com/labstack/echo/v4"
)

func (w WWW) Post() {
	tmpl := &Template{
		templates: template.Must(template.ParseFiles(projectDirectory() + "/templates/entry.html")),
	}
	w.E.Renderer = tmpl

	w.E.GET("post/entry/:id", func(c echo.Context) error {
		archive_id := server.ValidateArchiveID(*w.A, c.Param("id"))
		if archive_id == -1 {
			w.Serve404(c)
			return server.ErrInvalidArchiveID
		}

		/*
			file, err := w.A.GetFile(context.TODO(), archive_id)
			if err != nil {
				return err
			}
		*/
		tags, err := w.A.GetTags(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		hashes, err := w.A.GetHashes(context.TODO(), archive_id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		timestamps, err := w.A.GetTimestamps(context.TODO(), archive_id)
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
			},
		}); err != nil {
			fmt.Printf("error rendering Post. %v\n", err)
			w.Serve404(c)
			return err
		}

		return nil
	})
}

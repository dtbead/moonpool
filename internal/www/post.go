package www

import (
	"context"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"strconv"

	"github.com/dtbead/moonpool/internal/file"
	"github.com/labstack/echo/v4"
)

func (w WWW) Post() {
	w.echo.GET("post/entry/:id", func(c echo.Context) error {
		if w.config.DynamicWebReloading {
			tmp, err := template.ParseFiles(w.config.DynamicWebReloadingPath + "/templates/entry.html")
			if err != nil {
				return err
			}
			w.echo.Renderer = &Template{tmp}
		}

		ctx := context.Background()
		searchOptions := parseSearchOptions(c)

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			return fmt.Errorf("invalid archive ID")
		}

		media, err := w.api.GetPath(ctx, archive_id)
		if err != nil {
			return err
		}

		hashes, err := w.api.GetHashes(ctx, archive_id)
		if err != nil {
			return err
		}

		timestamps, err := w.api.GetTimestamps(ctx, archive_id)
		if err != nil {
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

		tags, err := w.api.GetTags(ctx, archive_id)
		if err != nil {
			return err
		}

		metadata, err := w.api.GetMetadata(ctx, archive_id)
		if err != nil {
			return err
		}

		pHash, _ := w.api.GetPerceptualHash(ctx, archive_id, "")

		if err := c.Render(http.StatusOK, "entry.html", map[string]interface{}{
			"archive_id":  archive_id,
			"searchQuery": searchOptions,
			"tagList":     tags,
			"hashes": map[string]string{
				"md5":    file.ByteToHexString(hashes.MD5),
				"sha1":   file.ByteToHexString(hashes.SHA1),
				"sha256": file.ByteToHexString(hashes.SHA256),
				"phash":  strconv.FormatUint(pHash, 16),
			},
			"timestamps": map[string]string{
				"imported": timeToString(timestamps.DateImported.Local()),
				"modified": timeToString(timestamps.DateModified.Local()),
				"created":  timeToString(timestamps.DateCreated.Local()),
			},
			"metadata": map[string]any{
				"filesize":          int64ToString(metadata.FileSize),
				"mimetype":          metadata.FileMimetype,
				"media_orientation": metadata.MediaOrientation,
				"height":            int64ToString(metadata.MediaHeight),
				"width":             int64ToString(metadata.MediaWidth),
			},
			"media":     media.FileRelative,
			"extension": mime.TypeByExtension(media.FileExtension),
		}); err != nil {
			fmt.Printf("error rendering post. %v\n", err)
			return err
		}
		return nil
	})
}

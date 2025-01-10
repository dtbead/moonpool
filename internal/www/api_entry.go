package www

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/importer"
	"github.com/dtbead/moonpool/internal/file"
	"github.com/labstack/echo/v4"
)

// replaceTags replaces all tags associated with a given archive_id
func (w WWW) replaceTags() {
	w.echo.POST("api/entry/:id/tags/replace", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"error": "post not found"})
			return errors.New("invalid archive id")
		}

		tags := c.FormValue("tags")
		if tags == "" {
			c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "no tags given"})
			return nil
		}

		if err := w.api.ReplaceTags(ctx, archive_id, strings.Split(tags, "\n")); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "unable to replace tags"})
			return err
		}

		c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
		return nil
	})
}

// entry returns all associated metadata with a given archive_id
func (w WWW) entry() {
	w.echo.GET("api/entry/:id", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		path, err := w.api.GetPath(ctx, archive_id)
		if err != nil {
			fmt.Printf("[%s] ERROR: failed to get path for archive id %d. %v\n", c.Request().RemoteAddr, archive_id, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
		}

		hashes, err := w.api.GetHashes(ctx, archive_id)
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to get hashes for archive id %d. %v\n", c.Request().RemoteAddr, archive_id, err)
		}

		tags, err := w.api.GetTags(ctx, archive_id)
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to get tags for archive id %d. %v\n", c.Request().RemoteAddr, archive_id, err)
		}

		timestamps, err := w.api.GetTimestamps(ctx, archive_id)
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to get timestamps for archive id %d. %v\n", c.Request().RemoteAddr, archive_id, err)
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"archive_id": archive_id,
			"extension":  path.FileExtension,
			"timestamps": map[string]string{
				"date_created":  timestamps.DateCreated.String(),
				"date_modified": timestamps.DateModified.String(),
				"date_imported": timestamps.DateImported.String(),
			},
			"hashes": map[string]string{
				"md5":    file.ByteToHexString(hashes.MD5),
				"sha1":   file.ByteToHexString(hashes.SHA1),
				"sha256": file.ByteToHexString(hashes.SHA256),
			},
			"tags": tags,
		})

		fmt.Printf("[%s] INFO: sent post %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

// removeTags unassigns all tags associated with a given archive_id
func (w WWW) removeTags() {
	w.echo.DELETE("api/entry/:id/tags", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		var tags []string
		if err := c.Bind(&tags); err != nil {
			fmt.Printf("[%s] WARNING: unable to parse tag request. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid json tags"})
			return err
		}

		if err := w.api.RemoveTags(ctx, archive_id, tags); err != nil {
			fmt.Printf("[%s] WARNING: failed to remove tag. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "unable to remove tags"})
			return err
		}

		c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
		return nil
	})

}

func (w WWW) upload() {
	w.echo.POST("api/entry/upload", func(c echo.Context) error {
		formFile, err := c.FormFile("file")
		if err != nil {
			fmt.Printf("[%s] ERROR: unknown error during upload. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		if formFile.Size <= 512 {
			fmt.Printf("[%s] WARNING: recieved filesize smaller than 512 bytes. Got %d bytes\n", c.Request().RemoteAddr, formFile.Size)
			c.JSON(http.StatusRequestedRangeNotSatisfiable, map[string]interface{}{"message": "filesize too small"}) // TODO: set content-range header accordingly
			return errors.New("too small of filesize")
		}

		if formFile.Size >= 25*1000000 { // TODO: test if this check works
			fmt.Printf("[%s] WARNING: recieved filesize greater than 25 megabytes. Got %d megabytes\n", c.Request().RemoteAddr, formFile.Size*megabyte)
			c.JSON(http.StatusRequestEntityTooLarge, map[string]interface{}{"message": "filesize too large"})
			return errors.New("too large of filesize")
		}

		extension, err := mime.ExtensionsByType(formFile.Header.Get("Content-Type"))
		if err != nil || extension == nil {
			fmt.Printf("[%s] ERROR: unable to get extension from mimetype. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "bad content-type"})
			return err
		}

		multipartFile, err := formFile.Open()
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to open multipart file. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "bad file"})
			return err
		}

		entry, err := importer.New(multipartFile, extension[0])
		if err != nil {
			fmt.Printf("[%s] ERROR: failed to create new entry. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		archive_id, err := w.api.Import(context.Background(), entry) // TODO: get tags from upload request
		if err != nil {
			fmt.Printf("[%s] ERROR: failed to import. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		c.JSON(http.StatusAccepted, map[string]interface{}{"id": archive_id, "url": fmt.Sprintf("%s/post/entry/%d", c.Echo().Server.Addr, archive_id)})
		fmt.Printf("[%s] INFO: successful import for archive_id %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

// TODO: add support for DateCreated and DateImported timestamps
func (w WWW) setTimestamps() {
	w.echo.POST("api/entry/:id/timestamps", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		type timestamps struct {
			DateModified int64 `json:"date_modified"`
		}
		var ts timestamps

		if err := c.Bind(&ts); err != nil {
			fmt.Printf("[%s] WARNING: failed to bind timestamp. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid timestamp"})
			return err
		}

		if ts.DateModified <= 0 {
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "no timestamp given"})
			return errors.New("no timestamp given")
		}

		err := w.api.SetTimestamps(ctx, archive_id, entry.Timestamp{
			DateModified: time.Unix(ts.DateModified, 0),
		})
		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
			return err
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
	})
}

func (w WWW) getTimestamps() {
	w.echo.GET("api/entry/:id/timestamps", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		ts, err := w.api.GetTimestamps(ctx, archive_id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "id not found"})
			return err
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		c.JSON(http.StatusAccepted, ts)
		return nil
	})
}

func (w WWW) getHashes() {
	w.echo.GET("api/entry/:id/hashes", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		hashes, err := w.api.GetHashes(ctx, archive_id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "id not found"})
			return err
		}
		defer isDeadlined(c, err)

		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, hashes)
	})
}

func (w WWW) getFile() {
	w.echo.GET("api/entry/:id/file", func(c echo.Context) error {
		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}
		entry, err := w.api.GetPath(context.TODO(), archive_id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{"error": "post not found"})
			return err
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "unknown error"})
			return err
		}

		fullPath := w.api.Config.MediaLocation + "/" + entry.FileRelative
		if err := c.File(fullPath); err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		return nil
	})
}

func (w WWW) deleteEntry() {
	w.echo.DELETE("api/entry/:id", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		err := w.api.RemoveArchive(ctx, archive_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
	})
}

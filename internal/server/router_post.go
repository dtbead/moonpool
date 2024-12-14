package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/importer"
	"github.com/dtbead/moonpool/internal/file"
	"github.com/labstack/echo/v4"
)

const megabyte = 1000000

// Post returns the metadata of a entry
func (s Server) post() {
	s.e.GET("post/entry/:id", func(c echo.Context) error {
		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		path, err := s.a.GetPath(context.TODO(), archive_id)
		if err != nil {
			fmt.Printf("[%s] ERROR: failed to get path for archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
		}

		hashes, err := s.a.GetHashes(context.TODO(), archive_id)
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to get hashes for archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
		}

		tags, err := s.a.GetTags(context.TODO(), archive_id)
		if err != nil {
			fmt.Printf("[%s] WARNING: failed to get tags for archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"archive_id": archive_id,
			"extension":  path.FileExtension,
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

func (s Server) setTags() {
	s.e.POST("post/set_tags/:id", func(c echo.Context) error {
		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		var tags []string
		if err := c.Bind(&tags); err != nil {
			fmt.Printf("[%s] WARNING: unable to parse tag request. %v", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid json tags"})
			return err
		}

		if tags == nil {
			fmt.Printf("[%s] INFO: received no tags to set for archive id %d.", c.Request().RemoteAddr, archive_id)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "no tags given"})
			return errors.New("no tags recieved")
		}

		if err := s.a.SetTags(context.Background(), archive_id, tags); err != nil {
			fmt.Printf("[%s] ERROR: failed to set tag on archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unable to set tags"})
			return err
		}

		// TODO: return list of tags of how the server decided to process them
		c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
		return nil
	})
}

func (s Server) removeTags() {
	s.e.POST("post/remove_tags/:id", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		var tags []string
		if err := c.Bind(&tags); err != nil {
			fmt.Printf("[%s] WARNING: unable to parse tag request. %v", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid json tags"})
			return err
		}

		if err := s.a.RemoveTags(ctx, archive_id, tags); err != nil {
			fmt.Printf("[%s] WARNING: failed to remove tag. %v", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "unable to remove tags"})
			return err
		}

		c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
		return nil
	})

}

func (s Server) upload() {
	s.e.POST("post/upload", func(c echo.Context) error {
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

		if formFile.Size >= 25*megabyte { // TODO: test if this check works
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

		archive_id, err := s.a.Import(context.Background(), entry) // TODO: get tags from upload request
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
func (s Server) setTimestamps() {
	s.e.POST("post/set_timestamps/:id", func(c echo.Context) error {
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
			fmt.Printf("[%s] WARNING: received no timestamp.\n", c.Request().RemoteAddr)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "no timestamp given"})
			return errors.New("no timestamp given")
		}

		err := s.a.SetTimestamps(ctx, archive_id, entry.Timestamp{
			DateModified: time.Unix(ts.DateModified, 0),
		})
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("[%s] WARNING: request timed-out\n", c.Request().RemoteAddr)
			c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
			return err
		}
		if err != nil {
			fmt.Printf("[%s] ERROR: failed to set timestamp. %v. timestamp = %v\n", c.Request().RemoteAddr, err, ts)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, map[string]interface{}{"message": "success"})
	})
}

func (s Server) getTimestamps() {
	s.e.GET("post/get_timestamps/:id", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		ts, err := s.a.GetTimestamps(ctx, archive_id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "id not found"})
			return err
		}

		if err != nil {
			fmt.Printf("[%s] ERROR: failed to get timestamp for archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		c.JSON(http.StatusAccepted, ts)
		return nil
	})
}

func (s Server) search() {
	s.e.POST("post/search", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		var tags []string
		if err := c.Bind(&tags); err != nil {
			fmt.Printf("[%s] WARNING: failed to bind tags. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid tags"})
			return err
		}

		res, err := s.a.SearchTag(ctx, tags[0])
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("[%s] WARNING: request timed-out to search tags\n", c.Request().RemoteAddr)
			c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
			return err
		}

		if err != nil {
			fmt.Printf("[%s] WARNING: failed to search tags. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, res)
	})
}

func (s Server) getHashes() {
	s.e.GET("post/get_hashes/:id", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}

		hashes, err := s.a.GetHashes(ctx, archive_id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "id not found"})
			return err
		}
		defer isDeadlined(c, err)

		if err != nil {
			fmt.Printf("[%s] ERROR: failed to get hashes on archive id %d. %v", c.Request().RemoteAddr, archive_id, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		return c.JSON(http.StatusAccepted, hashes)
	})
}

func (s Server) getFile() {
	s.e.GET("post/get_file/:id", func(c echo.Context) error {
		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			c.JSON(http.StatusNotFound, map[string]interface{}{"message": "post not found"})
			return errors.New("invalid archive id")
		}
		entry, err := s.a.GetPath(context.TODO(), archive_id)
		if err == sql.ErrNoRows {
			fmt.Printf("[%s] INFO: unable to find post = %d\n", c.Request().RemoteAddr, archive_id)
			c.JSON(http.StatusNotFound, map[string]interface{}{"error": "post not found"})
			return err
		}

		if err != nil {
			fmt.Printf("[%s] WARNING: unable to fulfil request. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "unknown error"})
			return err
		}

		fullPath := s.a.Config.MediaLocation + "/" + entry.FileRelative
		if err := c.File(fullPath); err != nil {
			fmt.Printf("[%s]\tWARNING: unable to retrieve file. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		fmt.Printf("[%s]\tINFO: sent file %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

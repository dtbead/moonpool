package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strconv"

	"github.com/dtbead/moonpool/archive"
	"github.com/labstack/echo/v4"
)

const MEGABYTE = 10000000

// Post returns the metadata of a entry
func (m Moonpool) Post() {
	m.E.GET("post/:id", func(c echo.Context) error {
		archive_id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			fmt.Printf("[%s] WARNING: received invalid archive id request\n", c.Request().RemoteAddr)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "invalid id"})
			return err
		}
		entry, err := m.a.Get(context.Background(), archive_id)
		if err != nil {
			fmt.Printf("[%s]\tWARNING: unable to fulfil request. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"archive_id": archive_id,
			"metadata": map[string]string{
				"md5":    byteToHex(entry.Hash().MD5),
				"sha1":   byteToHex(entry.Hash().SHA1),
				"sha256": byteToHex(entry.Hash().SHA256),
			},
			"tags": entry.Tags,
		})

		fmt.Printf("[%s]\tWARNING: sent post %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

func (m Moonpool) GetFile() {
	m.E.GET("getfile/:id", func(c echo.Context) error {
		archive_id := m.parseArchiveID(c.Param("id"))
		if archive_id < 1 {
			fmt.Printf("[%s] WARNING: received invalid post id request\n", c.Request().RemoteAddr)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid id"})
			return errors.New("invalid archive id")
		}

		entry, err := m.a.Get(context.Background(), archive_id)
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

		if err := c.File(entry.Path()); err != nil {
			fmt.Printf("[%s]\tWARNING: unable to retrieve file. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		fmt.Printf("[%s]\tINFO: sent file %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

func (m Moonpool) Upload() {
	m.E.POST("upload", func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Printf("[%s]\tERROR: unknown error during upload. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		if file.Size <= 512 {
			fmt.Printf("[%s]\tWARNING: recieved filesize smaller than 512 bytes. Got %d bytes\n", c.Request().RemoteAddr, file.Size)
			c.JSON(http.StatusRequestedRangeNotSatisfiable, map[string]interface{}{"message": "filesize too small"}) // TODO: set content-range header accordingly
			return errors.New("too small of filesize")
		}

		if file.Size == 25*MEGABYTE { // TODO: fix this; doesn't calculate filesize correctly
			fmt.Printf("[%s]\tWARNING: recieved filesize greater than 25 megabytes. Got %d bytes\n", c.Request().RemoteAddr, file.Size)
			c.JSON(http.StatusRequestEntityTooLarge, map[string]interface{}{"message": "filesize too large"})
			return errors.New("too large of filesize")
		}

		reader, err := file.Open()
		if err != nil {
			fmt.Printf("[%s]\tERROR: unable to open uploaded file. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		extension, err := mime.ExtensionsByType(file.Header.Get("Content-Type"))
		if err != nil {
			fmt.Printf("[%s]\tERROR: unable to get extension from mimetype. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "bad Content-Type"})
			return err
		}

		if extension == nil {
			fmt.Printf("[%s]\tERROR: unknown extension from Content-Type %v\n", c.Request().RemoteAddr, file.Header.Get("Content-Type"))
			c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "bad Content-Type"})
			return err
		}

		entry, err := archive.New(reader, extension[0])
		if err != nil {
			fmt.Printf("[%s]\tERROR: failed to create new entry. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "unknown error"})
			return err
		}

		archive_id, err := m.a.Import(context.Background(), entry, nil) // TODO: get tags from upload request
		if err != nil {
			fmt.Printf("[%s]\tERROR: failed to import. %v\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusFound, map[string]interface{}{"message": "unknown error"})
			return err
		}

		c.JSON(http.StatusAccepted, map[string]interface{}{"id": archive_id, "url": fmt.Sprintf("%s/post/%d", c.Echo().Server.Addr, archive_id)})
		return nil
	})
}

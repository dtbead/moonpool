package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

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
			"tags": entry.Tags})
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

		if err := c.File(entry.Metadata.PathRelative); err != nil {
			fmt.Printf("[%s]\tWARNING: unable to retrieve file. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"archive_id": archive_id,
			"metadata": map[string]string{
				"md5":    byteToHex(entry.Hash().MD5),
				"sha1":   byteToHex(entry.Hash().SHA1),
				"sha256": byteToHex(entry.Hash().SHA256),
			},
			"tags": entry.Tags})
		return nil
	})
}

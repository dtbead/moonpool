package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (m Moonpool) GetFile() {
	m.E.GET("post/get_file/:id", func(c echo.Context) error {
		archive_id := ValidateArchiveID(*m.A, c.Param("id"))
		if archive_id < 1 {
			fmt.Printf("[%s] WARNING: received invalid post id request\n", c.Request().RemoteAddr)
			c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid id"})
			return ErrInvalidArchiveID
		}

		entry, err := m.A.GetPath(context.TODO(), archive_id)
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

		if err := c.File(entry.Filepath); err != nil {
			fmt.Printf("[%s]\tWARNING: unable to retrieve file. %s\n", c.Request().RemoteAddr, err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": "failed to retrieve content"})
			return err
		}

		fmt.Printf("[%s]\tINFO: sent file %d\n", c.Request().RemoteAddr, archive_id)
		return nil
	})
}

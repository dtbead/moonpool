package www

import (
	"context"

	"github.com/dtbead/moonpool/internal/server"
	"github.com/labstack/echo/v4"
)

func (w WWW) Thumbnail() {
	w.echo.GET("thumbnail/:id", func(c echo.Context) error {
		archive_id := server.ValidateArchiveID(*w.api, c.Param("id"))
		if archive_id == -1 {
			return server.ErrInvalidArchiveID
		}

		thumb, err := w.api.GetThumbnail(context.Background(), archive_id, "small", "webp")
		if err != nil {
			return err
		}

		if _, err := c.Response().Write(thumb); err != nil {
			return err
		}
		c.Response().Header().Add("Content-Type", "image/webp")
		c.Response().Flush()

		return nil
	})
}

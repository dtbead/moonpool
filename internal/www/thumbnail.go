package www

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
)

func (w WWW) Thumbnail() {
	w.echo.GET("thumbnail/:id", func(c echo.Context) error {
		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			return errors.New("invalid archive id")
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

package www

import (
	_ "embed"
	"errors"

	"github.com/dtbead/moonpool/api"
	"github.com/labstack/echo/v4"
)

//go:embed web/assets/static/404.png
var defaultThumbnailPNG []byte

func (w WWW) Thumbnail() {
	w.echo.GET("thumbnail/:id", func(c echo.Context) error {
		defaultThumbnail := func(c echo.Context) {
			c.Response().Write(defaultThumbnailPNG)
			c.Response().Header().Add("Content-Type", "image/jpeg")
			c.Response().Flush()
		}

		archive_id := stringToInt64(c.Param("id"))
		if archive_id <= 0 {
			defaultThumbnail(c)
			return errors.New("invalid archive id")
		}

		thumb, err := w.api.GetThumbnail(c.Request().Context(), archive_id, "small", "jpeg")
		if errors.Is(err, api.ErrThumbnailNotFound) {
			defaultThumbnail(c)
			return nil
		}

		if err != nil {
			defaultThumbnail(c)
			return err
		}

		c.Response().Header().Add("Content-Type", "image/jpeg")
		if _, err := c.Response().Write(thumb); err != nil {
			return err
		}

		c.Response().Flush()
		return nil
	})
}

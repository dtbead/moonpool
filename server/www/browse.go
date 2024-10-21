package www

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (w WWW) Browse() {
	w.echo.GET("browse", func(c echo.Context) error {
		if err := c.Render(http.StatusOK, "browse.html", nil); err != nil {
			fmt.Printf("error rendering Post. %v\n", err)
			return err
		}
		return nil
	})
}

package www

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const dateFormat = "Jan 2 2006, 3:04:05 PM"
const megabyte = 1000000

func timeToString(t time.Time) string {
	return t.Format(dateFormat)
}

// returns -1 if given invalid string
func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return i
}

func int64ToString(i int64) string {
	return strconv.FormatInt(int64(i), 10)
}

func add(n, s int64) int64 {
	if n+s <= 0 {
		return 0
	}
	return n + s
}

func isDeadlined(c echo.Context, err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("[%s] WARNING: request timed-out\n", c.Request().RemoteAddr)
		c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
		return c.JSON(http.StatusRequestTimeout, map[string]interface{}{"message": "request took too long to complete"})
	}
	return nil
}

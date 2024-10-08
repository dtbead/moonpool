package thumbnail

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/dtbead/moonpool/entry"
)

type Thumbnail struct {
	jpeg, webp entry.Icons
}

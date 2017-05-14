package images

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func Size(file string) (width int, height int, err error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, 0, fmt.Errorf("opening image file %q: %v", file, err)
	}
	defer f.Close()
	c, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("decoding image file %q: %v", file, err)
	}
	return c.Width, c.Height, nil
}

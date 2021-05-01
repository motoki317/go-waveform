package waveform

import (
	"fmt"
	"image/color"
)

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf(
		"#%02x%02x%02x",
		r>>8,
		g>>8,
		b>>8,
	)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

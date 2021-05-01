package waveform

import (
	"fmt"
	"image/color"
)

const (
	gradMinAlpha = 0x80
	gradMaxAlpha = 0xc0
	gradStep     = 1
	gradLoop     = (gradMaxAlpha - gradMinAlpha) / gradStep
)

func gradation(c color.Color, i int) color.Color {
	r, g, b, _ := c.RGBA()
	i %= gradLoop * 2
	var a uint8
	if i < gradLoop {
		a = uint8(gradMinAlpha + i*gradStep)
	} else {
		a = uint8(gradMaxAlpha - (i-gradLoop)*gradStep)
	}
	fmt.Println(a)
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: a,
	}
}

func colorToHex(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf(
		"#%02x%02x%02x%02x",
		r>>8,
		g>>8,
		b>>8,
		a>>8,
	)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

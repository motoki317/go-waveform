package waveform

import (
	"image/color"
)

// Option is an image generation option.
type Option struct {
	// Resolution specifies the resolution of the image.
	// Required.
	Resolution int
	// Width specifies the width of the resulting image.
	// Default: Resolution * 5
	Width int
	// Height specifies the height of the resulting image.
	// Default: 540
	Height int
	// Background specifies the color of the background.
	// Default: none (transparent)
	Background color.Color
	// Color specifies the color of the waveform.
	// Default: color.Black
	Color color.Color
}

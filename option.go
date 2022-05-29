package waveform

import (
	"errors"
	"image/color"
)

// Option is an image generation option.
type Option struct {
	// Resolution specifies the number of bars in the image.
	// Required.
	Resolution int
	// Width specifies the width of the resulting image in pixels.
	// Default: Resolution * 5 px
	Width int
	// Height specifies the height of the resulting image in pixels.
	// Default: 540 px
	Height int
	// Background specifies the color of the background.
	// Default: nil (transparent)
	Background color.Color
	// Color specifies the color of the waveform bars.
	// Default: color.Black
	Color color.Color
}

func (o *Option) validate() error {
	if o.Resolution == 0 {
		return errors.New("resolution option is required")
	}
	return nil
}

func (o *Option) applyDefaults(sampleLength int) {
	// Corner case: if sample length is less than resolution, make resolution equal to sample length
	if sampleLength < o.Resolution {
		o.Resolution = sampleLength
	}
	if o.Width == 0 {
		o.Width = o.Resolution * 5
	}
	if o.Height == 0 {
		o.Height = 540
	}
	if o.Color == nil {
		o.Color = color.Black
	}
}

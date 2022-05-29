package waveform

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"math"
	"time"

	svg "github.com/ajstarks/svgo/float"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"

	"github.com/motoki317/go-waveform/internal/reader"
	"github.com/motoki317/go-waveform/internal/utils"
)

// bound sample value upper and lower boundary
type bound struct {
	Upper float64
	Lower float64
}

type svgWriter struct {
	s            *svg.SVG
	reader       reader.Reader
	sampleLength int
	bound        *bound
	option       *Option
}

func (s *svgWriter) write() error {
	n := s.option.Resolution
	batchRead := int(float64(s.sampleLength)/float64(n) + 0.5)
	if n > s.sampleLength {
		n = s.sampleLength
		batchRead = 1
	}

	rectWidth := float64(2)
	width := float64(n * 5)
	height := 540.
	if s.option.Width != 0 {
		width = float64(s.option.Width)
		rectWidth = width / float64(n) * 0.4
	}
	if s.option.Height > 0 {
		height = float64(s.option.Height)
	}

	s.s.Start(width, height)
	if s.option.Background != nil {
		s.s.Rect(0, 0, width, height, `fill="`+utils.ColorToHex(s.option.Background)+`"`)
	}

	floor := (s.bound.Upper + s.bound.Lower) / 2
	sampleHeight := s.bound.Upper - s.bound.Lower
	lineCol := s.option.Color
	if lineCol == nil {
		lineCol = color.Black
	}

	i := 0
	buf := make([]float64, batchRead)
	for i < s.sampleLength {
		expectBytes := utils.Min(batchRead, s.sampleLength-i)
		read, err := s.reader.ReadNSamples(buf[:expectBytes])
		if err != nil {
			return err
		}
		if len(read) == 0 {
			break
		}
		min, max := utils.GetMinMax(floor, read)

		x := float64(i) / float64(s.sampleLength) * width
		y := (min - s.bound.Lower) / sampleHeight * height
		h := (max - min) / sampleHeight * height
		s.s.Rect(x, y, rectWidth, h, `fill="`+utils.ColorToHex(lineCol)+`"`)

		i += len(read)
	}

	return nil
}

func outputWaveformImage(sample reader.Reader, sampleLength int, bound *bound, option *Option) (r io.Reader, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered: %v", r)
			}
		}
	}()

	b := bytes.NewBuffer(make([]byte, 0))

	s := svg.New(b)
	writer := &svgWriter{
		s:            s,
		reader:       sample,
		sampleLength: sampleLength,
		bound:        bound,
		option:       option,
	}
	if err = writer.write(); err != nil {
		return
	}
	writer.s.End()

	return b, nil
}

// OutputWaveformImageMp3 outputs waveform image from *mp3.Decoder.
func OutputWaveformImageMp3(data *mp3.Decoder, option *Option) (r io.Reader, err error) {
	d := reader.NewMp3Decoder(data)
	return outputWaveformImage(d, int(data.Length()/4), &bound{
		Upper: 32767,
		Lower: -32768,
	}, option)
}

// OutputWaveformImageWav outputs waveform image from *wav.Decoder.
func OutputWaveformImageWav(data *wav.Decoder, option *Option) (r io.Reader, err error) {
	d := reader.NewWavDecoder(data)
	data.ReadInfo()
	dur, err := data.Duration()
	if err != nil {
		return nil, err
	}
	byteLen := int((float64(dur) * float64(data.AvgBytesPerSec)) / float64(time.Second))
	if data.BitDepth < 8 || data.NumChans == 0 {
		return nil, fmt.Errorf("failed to retrieve correct bit depth / num channels (%d, %d)", data.BitDepth, data.NumChans)
	}
	return outputWaveformImage(d, byteLen/int(data.BitDepth/8)/int(data.NumChans), &bound{
		Upper: math.Pow(2, float64(data.BitDepth-1)) - 1,
		Lower: -math.Pow(2, float64(data.BitDepth-1)),
	}, option)
}

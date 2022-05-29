package waveform

import (
	"bytes"
	"fmt"
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
	res := float64(s.option.Resolution)
	width := float64(s.option.Width)
	height := float64(s.option.Height)

	batchRead := int(float64(s.sampleLength)/res + 0.5)
	rectWidth := width / res * 0.4 /* 40% of the width */
	sampleHeight := s.bound.Upper - s.bound.Lower

	s.s.Start(width, height)
	if s.option.Background != nil {
		s.s.Rect(0, 0, width, height, `fill="`+utils.ColorToHex(s.option.Background)+`"`)
	}

	readSamples := 0
	buf := make([]float64, batchRead)
	for readSamples < s.sampleLength {
		expectBytes := utils.Min(batchRead, s.sampleLength-readSamples)
		read, err := s.reader.ReadNSamples(buf[:expectBytes])
		if err != nil {
			return err
		}
		if len(read) == 0 {
			break
		}

		x := float64(readSamples) / float64(s.sampleLength) * width
		// Normalize samples before passing to BarDrawer
		// [s.bound.Lower, s.bound.Upper] -> [-1, 1]
		for i := range read {
			read[i] = (read[i]-s.bound.Lower)/sampleHeight*2 - 1
		}
		y, h := s.option.Drawer(read)
		y *= height
		h *= height
		s.s.Rect(x, y, rectWidth, h, `fill="`+utils.ColorToHex(s.option.Color)+`"`)

		readSamples += len(read)
	}

	return nil
}

func outputWaveformImage(reader reader.Reader, sampleLength int, bound *bound, option *Option) (r io.Reader, err error) {
	if err := option.validate(); err != nil {
		return nil, err
	}
	option.applyDefaults(sampleLength)

	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered: %v", r)
			}
		}
	}()

	var b bytes.Buffer
	writer := &svgWriter{
		s:            svg.New(&b),
		reader:       reader,
		sampleLength: sampleLength,
		bound:        bound,
		option:       option,
	}
	if err = writer.write(); err != nil {
		return
	}
	writer.s.End()

	return &b, nil
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

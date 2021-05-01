package waveform

import (
	"errors"
	"fmt"
	svg "github.com/ajstarks/svgo/float"
	"image/color"
	"io"
	"math"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
)

// Option image option
type Option struct {
	// Resolution specifies the resolution of the
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

// bound sample value upper and lower boundary
type bound struct {
	Upper float64
	Lower float64
}

type float64Reader interface {
	// readNSamples reads the next n samples, and return the values in float64 slice.
	readNSamples(buf []float64) ([]float64, error)
}

type mp3Decoder struct {
	*mp3.Decoder
	buf []byte
}

func (d *mp3Decoder) readNSamples(buf []float64) ([]float64, error) {
	if d.buf == nil || cap(d.buf) < len(buf)*4 {
		d.buf = make([]byte, len(buf)*4)
	}

	totalSamples := 0
	for totalSamples < len(buf) {
		expectBytes := len(d.buf) - totalSamples*4
		read, err := d.Read(d.buf[:expectBytes])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("an error occurred while decoding mp3: %w", err)
		}
		if err == io.EOF {
			break
		}
		if read%4 != 0 {
			return nil, errors.New("expected multiple of 4 bytes to be read")
		}

		// 16 bit 2 channels
		src := d.buf[:read]
		for i := 0; i < read/4; i++ {
			buf[totalSamples+i] = float64(int16(uint16(src[i*4]) | uint16(src[i*4+1])<<8))
		}

		totalSamples += read / 4
	}
	return buf[:totalSamples], nil
}

type wavDecoder struct {
	*wav.Decoder
	buf *audio.IntBuffer
}

func (d *wavDecoder) readNSamples(buf []float64) ([]float64, error) {
	numCh := int(d.NumChans)
	if d.buf == nil {
		d.buf = &audio.IntBuffer{
			Data: make([]int, len(buf)*numCh),
		}
	}

	read, err := d.PCMBuffer(d.buf)
	if err != nil {
		return nil, err
	}

	src := d.buf.Data[:read]
	for i := 0; i < read/numCh; i++ {
		buf[i] = float64(src[i*numCh])
	}
	return buf[:read/numCh], nil
}

type svgWriter struct {
	s            *svg.SVG
	sample       float64Reader
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
		s.s.Rect(0, 0, width, height, "fill:"+colorToHex(s.option.Background))
	}

	floor := (s.bound.Upper + s.bound.Lower) / 2
	sampleHeight := s.bound.Upper - s.bound.Lower
	lineCol := s.option.Color
	if lineCol == nil {
		lineCol = color.Black
	}

	i := 0
	lineNum := 0
	buf := make([]float64, batchRead)
	for i < s.sampleLength {
		expectBytes := min(batchRead, s.sampleLength-i)
		read, err := s.sample.readNSamples(buf[:expectBytes])
		if err != nil {
			return err
		}
		if len(read) == 0 {
			break
		}
		min, max := getMinMax(floor, read)

		x := float64(i) / float64(s.sampleLength) * width
		y := (min - s.bound.Lower) / sampleHeight * height
		h := (max - min) / sampleHeight * height
		s.s.Rect(x, y, rectWidth, h, "fill:"+colorToHex(gradation(lineCol, lineNum)))

		i += len(read)
		lineNum++
	}

	return nil
}

func outputWaveformImage(sample float64Reader, sampleLength int, bound *bound, option *Option) (io.Reader, error) {
	r, w := io.Pipe()
	go func() {
		s := svg.New(w)
		writer := &svgWriter{
			s:            s,
			sample:       sample,
			sampleLength: sampleLength,
			bound:        bound,
			option:       option,
		}
		_ = writer.write()
		writer.s.End()
		_ = w.Close()
	}()
	return r, nil
}

func getMinMax(floor float64, s []float64) (min, max float64) {
	max, min = floor, floor

	for _, y := range s {
		if y > floor {
			if y > max {
				max = y
			}
		} else {
			if y < min {
				min = y
			}
		}
	}

	return
}

// OutputWaveformImageMp3 outputs waveform image from *mp3.Decoder.
func OutputWaveformImageMp3(data *mp3.Decoder, option *Option) (io.Reader, error) {
	d := &mp3Decoder{
		Decoder: data,
	}
	return outputWaveformImage(d, int(data.Length()/4), &bound{
		Upper: 32767,
		Lower: -32768,
	}, option)
}

// OutputWaveformImageWav outputs waveform image from *wav.Decoder.
func OutputWaveformImageWav(data *wav.Decoder, option *Option) (io.Reader, error) {
	d := &wavDecoder{
		Decoder: data,
	}
	data.ReadInfo()
	dur, err := data.Duration()
	if err != nil {
		return nil, err
	}
	byteLen := int((float64(dur) * float64(data.AvgBytesPerSec)) / float64(time.Second))
	return outputWaveformImage(d, byteLen/int(data.BitDepth/8)/int(data.NumChans), &bound{
		Upper: math.Pow(2, float64(data.BitDepth-1)) - 1,
		Lower: -math.Pow(2, float64(data.BitDepth-1)),
	}, option)
}

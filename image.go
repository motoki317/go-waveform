package waveform

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Option image option
type Option struct {
	FileName   string
	FileType   string
	Resolution int
	Width      int
	Theme      string
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

// OutputWaveformImageMp3 outputs waveform image from *mp3.Decoder.
func OutputWaveformImageMp3(data *mp3.Decoder, option *Option) error {
	d := &mp3Decoder{
		Decoder: data,
	}
	return outputWaveformImage(d, int(data.Length()/4), &bound{
		Upper: 32767,
		Lower: -32768,
	}, option, "")
}

// OutputWaveformImageWav outputs waveform image from *wav.Decoder.
func OutputWaveformImageWav(data *wav.Decoder, option *Option) error {
	d := &wavDecoder{
		Decoder: data,
	}
	data.ReadInfo()
	dur, err := data.Duration()
	if err != nil {
		return err
	}
	byteLen := int((float64(dur) * float64(data.AvgBytesPerSec)) / float64(time.Second))
	return outputWaveformImage(d, byteLen/int(data.BitDepth/8)/int(data.NumChans), &bound{
		Upper: math.Pow(2, float64(data.BitDepth-1)) - 1,
		Lower: -math.Pow(2, float64(data.BitDepth-1)),
	}, option, "")
}

func outputWaveformImage(sample float64Reader, sampleLength int, bound *bound, option *Option, postfix string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	floor := (bound.Upper + bound.Lower) / 2

	n := option.Resolution
	m := int(float64(sampleLength)/float64(n) + 0.5)

	if n > sampleLength {
		n = sampleLength
		m = 1
	}

	stroke := float64(2)
	width := float64(n * 5)

	if option.Width != 0 {
		width = float64(option.Width)
		stroke = width / float64(n) * 0.4
	}

	i := 0
	d := 1
	g := 155

	buf := make([]float64, m)
	for i < sampleLength {
		expectBytes := min(m, sampleLength-i)
		read, err := sample.readNSamples(buf[:expectBytes])
		if err != nil {
			return err
		}
		if len(read) == 0 {
			break
		}
		xys := getXYs(i, read, floor)

		l, err := plotter.NewLine(xys)
		if err != nil {
			return err
		}

		l.LineStyle.Width = vg.Points(stroke)
		l.Color = &color.RGBA{R: 50, G: uint8(g), B: 240, A: 255}

		p.Add(l)

		g += d
		i += len(read)

		if g > 225 {
			g = 225 - 1
			d = -1
		} else if g < 155 {
			g = 156
			d = 1
		}
	}

	p.HideX()
	p.HideY()
	p.X.Min = 0
	p.X.Max = float64(sampleLength)
	p.Y.Min = bound.Lower
	p.Y.Max = bound.Upper
	p.BackgroundColor = getBackgroundColor(option.Theme)

	fileName := fmt.Sprintf("%s%s.%s", option.FileName, postfix, option.FileType)

	return p.Save(vg.Points(width), vg.Points(540), fileName)
}

func getXYs(x int, s []float64, floor float64) *plotter.XYs {
	max := floor
	min := floor

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

	return &plotter.XYs{
		{X: float64(x), Y: min},
		{X: float64(x), Y: max},
	}
}

func getBackgroundColor(theme string) color.Color {
	switch theme {
	case "dark":
		return color.Gray{Y: 16}
	default:
		return color.Gray{Y: 240}
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

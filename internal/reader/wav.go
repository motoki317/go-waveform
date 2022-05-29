package reader

import (
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/motoki317/go-waveform/internal/utils"
)

type wavDecoder struct {
	*wav.Decoder
	buf *audio.IntBuffer
}

func NewWavDecoder(decoder *wav.Decoder) Reader {
	return &wavDecoder{Decoder: decoder}
}

func (d *wavDecoder) ReadNSamples(buf []float64) ([]float64, error) {
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

	// workaround: https://github.com/motoki317/go-waveform/issues/1
	read = utils.Min(read, len(d.buf.Data))

	src := d.buf.Data[:read]
	for i := 0; i < read/numCh; i++ {
		buf[i] = float64(src[i*numCh])
	}
	return buf[:read/numCh], nil
}

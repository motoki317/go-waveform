package reader

import (
	"errors"
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3"
)

type mp3Decoder struct {
	*mp3.Decoder
	buf []byte
}

func NewMp3Decoder(decoder *mp3.Decoder) Reader {
	return &mp3Decoder{Decoder: decoder}
}

func (d *mp3Decoder) ReadNSamples(buf []float64) ([]float64, error) {
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

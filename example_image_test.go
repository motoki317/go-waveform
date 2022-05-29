package waveform_test

import (
	"io/ioutil"
	"os"

	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"

	"github.com/motoki317/go-waveform"
)

func ExampleOutputWaveformImageMp3() {
	mp3File, err := os.Open(os.Getenv("FILENAME_IN_MP3"))
	if err != nil {
		panic(err)
	}

	d, err := mp3.NewDecoder(mp3File)
	if err != nil {
		panic(err)
	}
	r, err := waveform.OutputWaveformImageMp3(
		d,
		&waveform.Option{
			Resolution: 320,
		},
	)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(os.Getenv("FILENAME_OUT_MP3"), b, 0644)
	if err != nil {
		panic(err)
	}
}

func ExampleOutputWaveformImageWav() {
	wavFile, err := os.Open(os.Getenv("FILENAME_IN_WAV"))
	if err != nil {
		panic(err)
	}

	r, err := waveform.OutputWaveformImageWav(
		wav.NewDecoder(wavFile),
		&waveform.Option{
			Resolution: 320,
		},
	)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(os.Getenv("FILENAME_OUT_WAV"), b, 0644)
	if err != nil {
		panic(err)
	}
}

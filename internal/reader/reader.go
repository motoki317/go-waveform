package reader

type Reader interface {
	// ReadNSamples reads the next n samples, and return the values in float64 slice.
	ReadNSamples(buf []float64) ([]float64, error)
}

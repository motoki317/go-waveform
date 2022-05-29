package waveform

// BarDrawer determines the bar position and height.
// samples is a sequence of samples normalized in [-1, 1].
// BarDrawer should return y, h in range of [0, 1].
type BarDrawer func(samples []float64) (y, h float64)

// DrawerMinMax is a dead simple bar drawer.
// It extracts min/max values from the samples and determines position and height.
var DrawerMinMax BarDrawer = func(samples []float64) (y, h float64) {
	var min, max float64
	for _, s := range samples {
		if 0 <= s {
			if max < s {
				max = s
			}
		} else {
			if s < min {
				min = s
			}
		}
	}

	// [-1, 1] -> [0, 1]
	min = (min + 1) / 2
	max = (max + 1) / 2
	return min, max - min
}

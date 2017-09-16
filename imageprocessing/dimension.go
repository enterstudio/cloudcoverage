package imageprocessing

type ImageDimension struct {
	width, height int
}

func (i *ImageDimension) center() (float64, float64) {
	return float64(i.width) / 2, float64(i.height) / 2
}

package imagedimension

type ImageDimension struct {
	width, height int
}

var (
	FullSize    = ImageDimension{2592, 1944}
	HalfSize    = ImageDimension{FullSize.width / 2, FullSize.height / 2}
	QuarterSize = ImageDimension{HalfSize.width / 2, HalfSize.height / 2}
)

func (i *ImageDimension) Center() (float64, float64) {
	return float64(i.width) / 2, float64(i.height) / 2
}

func (i *ImageDimension) Width() int {
	return i.width
}

func (i *ImageDimension) Height() int {
	return i.height
}

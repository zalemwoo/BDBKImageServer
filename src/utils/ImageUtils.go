package utils

import (
	"image"

	"github.com/disintegration/gift"
)

func CropImage(img image.Image, rect image.Rectangle) (croped image.Image) {
	g := gift.New(
		gift.Crop(rect),
	)
	croped_image := image.NewRGBA(g.Bounds(rect))
	g.Draw(croped_image, img)
	return croped_image
}

package util

import "github.com/zmb3/spotify"

func FindSmallestImage(images []spotify.Image) (smallestImage spotify.Image) {
	// TODO: If no images are present, this does not work. Supply some fallback image in that case

	smallestImage = images[0]
	for _, currentImage := range images {
		if currentImage.Width*currentImage.Height < smallestImage.Width*smallestImage.Height {
			smallestImage = currentImage
		}
	}

	return smallestImage
}

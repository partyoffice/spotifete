package shared

import "github.com/zmb3/spotify"

func FindSmallestImageUrlOrEmpty(images []spotify.Image) (smallestImageUrl string) {
	smallestImage := findSmallestImage(images)
	if smallestImage == nil {
		return ""
	} else {
		return smallestImage.URL
	}
}

func findSmallestImage(images []spotify.Image) (smallestImage *spotify.Image) {
	if len(images) == 0 {
		return nil
	}

	smallestImage = &images[0]
	for _, currentImage := range images {
		if currentImage.Width*currentImage.Height < smallestImage.Width*smallestImage.Height {
			smallestImage = &currentImage
		}
	}

	return smallestImage
}

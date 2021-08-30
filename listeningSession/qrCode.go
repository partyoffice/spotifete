package listeningSession

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"net/http"

	"github.com/partyoffice/spotifete/config"
	. "github.com/partyoffice/spotifete/shared"
	"github.com/skip2/go-qrcode"
)

func QrCodeAsJpeg(joinId string, disableBorder bool, size int) (*bytes.Buffer, *SpotifeteError) {
	qrCode, spotifeteError := QrCode(joinId, disableBorder)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	jpegBuffer := new(bytes.Buffer)
	err := jpeg.Encode(jpegBuffer, qrCode.Image(size), nil)
	if err != nil {
		return nil, NewError("Could not encode qr code as image.", err, http.StatusInternalServerError)
	}

	return jpegBuffer, nil
}

func QrCodeAsPng(joinId string, disableBorder bool, size int) (*bytes.Buffer, *SpotifeteError) {
	qrCode, spotifeteError := QrCode(joinId, disableBorder)
	if spotifeteError != nil {
		return nil, spotifeteError
	}

	pngBuffer := new(bytes.Buffer)
	err := png.Encode(pngBuffer, qrCode.Image(size))
	if err != nil {
		return nil, NewError("Could not encode qr code as image.", err, http.StatusInternalServerError)
	}

	return pngBuffer, nil
}

func QrCode(joinId string, disableBorder bool) (qrcode.QRCode, *SpotifeteError) {
	qrCode, err := qrcode.New(qrCodeContent(joinId), qrcode.Medium)
	if err != nil {
		return qrcode.QRCode{}, NewError("Could not create QR code.", err, http.StatusInternalServerError)
	}

	qrCode.DisableBorder = disableBorder
	return *qrCode, nil
}

func qrCodeContent(joinId string) string {
	baseUrl := config.Get().SpotifeteConfiguration.BaseUrl
	return baseUrl + "/session/view/" + joinId
}

package utility

import (
	"image/color"

	qrcode "github.com/skip2/go-qrcode"
)

func GenerateQRCode(data string, backgroundColor, qrcodeColor color.Color) (qrcodeByte []byte, err error) {
	q, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	q.BackgroundColor = backgroundColor
	q.ForegroundColor = qrcodeColor

	return q.PNG(256)
}

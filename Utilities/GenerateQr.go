package Utilities

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/url"

	"github.com/skip2/go-qrcode"
)

const yandexShortenerURL = "https://clck.ru/--"

func ShortenURL(longURL string) (string, error) {
	encodedURL := url.QueryEscape(longURL)

	reqURL := fmt.Sprintf("https://clck.ru/--?url=%s", encodedURL)

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	shortURL, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("❌ Не удалось сократить ссылку")
	}

	return string(shortURL), nil
}

func GenerateQRCode(link string) ([]byte, error) {
	var qrCodeBuffer bytes.Buffer
	qrCode, err := qrcode.New(link, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	err = png.Encode(&qrCodeBuffer, qrCode.Image(256))
	if err != nil {
		return nil, err
	}

	return qrCodeBuffer.Bytes(), nil
}

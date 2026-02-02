// Package capture provides screen capture functionality.
package capture

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/kbinani/screenshot"
)

// CaptureAll captures the primary display and returns the image.
func CaptureAll() (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found")
	}
	return CaptureDisplay(0)
}

// CaptureDisplay captures a specific display by index.
func CaptureDisplay(index int) (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if index < 0 || index >= n {
		return nil, fmt.Errorf("display index %d out of range (0-%d)", index, n-1)
	}

	bounds := screenshot.GetDisplayBounds(index)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("failed to capture display %d: %w", index, err)
	}

	return img, nil
}

// EncodeToBase64JPEG encodes an image to base64 JPEG with specified quality.
func EncodeToBase64JPEG(img image.Image, quality int) (string, error) {
	if quality <= 0 || quality > 100 {
		quality = 80
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return "", fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// CaptureAndEncode captures the primary display and returns it as base64 JPEG.
func CaptureAndEncode() (string, error) {
	img, err := CaptureAll()
	if err != nil {
		return "", err
	}
	return EncodeToBase64JPEG(img, 80)
}

// GetDisplayCount returns the number of active displays.
func GetDisplayCount() int {
	return screenshot.NumActiveDisplays()
}

// GetDisplayBounds returns the bounds of a display.
func GetDisplayBounds(index int) (x, y, width, height int, err error) {
	n := screenshot.NumActiveDisplays()
	if index < 0 || index >= n {
		return 0, 0, 0, 0, fmt.Errorf("display index %d out of range (0-%d)", index, n-1)
	}

	bounds := screenshot.GetDisplayBounds(index)
	return bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy(), nil
}

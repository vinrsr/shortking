package qrcode

import "github.com/skip2/go-qrcode"

const DefaultSize = 256

// PNG renders content (typically a short link URL) as a PNG-encoded QR code.
// Generated on demand rather than persisted: it's cheap and deterministic
// from the input, so there's nothing to invalidate.
func PNG(content string, size int) ([]byte, error) {
	if size <= 0 {
		size = DefaultSize
	}
	return qrcode.Encode(content, qrcode.Medium, size)
}

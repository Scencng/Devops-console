package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func bindJSON(ctx *gin.Context, target any) error {
	if ctx.Request == nil || ctx.Request.Body == nil {
		return io.EOF
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	body = trimJSONLeadingNoise(body)
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

	decoder := json.NewDecoder(bytes.NewReader(body))
	return decoder.Decode(target)
}

func trimJSONLeadingNoise(body []byte) []byte {
	trimmed := body
	for len(trimmed) > 0 {
		switch {
		case bytes.HasPrefix(trimmed, utf8BOM):
			trimmed = trimmed[len(utf8BOM):]
			continue
		default:
			r, size := utf8.DecodeRune(trimmed)
			if r == utf8.RuneError && size == 1 {
				return bytes.TrimSpace(trimmed)
			}
			if unicode.IsSpace(r) {
				trimmed = trimmed[size:]
				continue
			}
		}
		break
	}

	return bytes.TrimSpace(trimmed)
}

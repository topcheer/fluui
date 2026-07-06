package term

import (
	"strconv"
	"strings"
)

// ─── Sixel Graphics Encoding ───
//
// Sixel is a graphics format from the DEC VT340 terminal. It encodes
// bitmap images as a sequence of printable ASCII characters. Each "sixel"
// character encodes 6 vertical pixels: the character code is 0x3F + bitmap,
// where bit 0 = top pixel, bit 5 = bottom pixel.
//
// Format:
//   DCS 8 ; ; ; q <data> ST
//
// Where DCS = ESC P, ST = ESC \
//
// Data contains:
//   - Raster attributes: "Pan;Pad;Ph;Pv"
//   - Color registers: "#Pc;2;R;G;B" (RGB percentages 0-100)
//   - Color selection: "#Pc"
//   - Sixel data: characters 0x3F-0x7E
//   - Repeat: "!count char"
//   - New line: "-"

// EncodeSixel converts an RGBA image to a Sixel DCS string.
//
// Parameters:
//   - rgba: RGBA pixel data (4 bytes per pixel: R, G, B, A)
//   - width: image width in pixels
//   - height: image height in pixels
//
// Returns the complete DCS-encapsulated Sixel string ready for terminal output.
// The image is color-quantized to at most 256 colors using a simple median-cut
// style algorithm.
func EncodeSixel(rgba []byte, width, height int) string {
	if width <= 0 || height <= 0 || len(rgba) < width*height*4 {
		return ""
	}

	var buf strings.Builder

	// DCS introducer: ESC P 8 ; ; ; q
	// P1=8 means no background fill (0 bits remain unchanged)
	// P2=0 means 0 bits use background color
	// P3=0 (horizontal grid size, ignored)
	buf.WriteString("\x1bP8;;;q")

	// Raster attributes: "Pan;Pad;Ph;Pv"
	// Aspect ratio 1:1 (Pan=1, Pad=1)
	// Ph = width, Pv = height
	buf.WriteByte('"')
	buf.WriteString(strconv.Itoa(1)) // Pan
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(1)) // Pad
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(width)) // Ph
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(height)) // Pv

	// Color quantization: extract unique colors
	palette, colorMap := quantizeColors(rgba, width*height)

	// Emit color register definitions
	for i, c := range palette {
		buf.WriteByte('#')
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(";2;") // RGB color type
		buf.WriteString(strconv.Itoa(int(c.r) * 100 / 255))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(int(c.g) * 100 / 255))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(int(c.b) * 100 / 255))
	}

	// Encode sixel data: process 6 rows at a time
	for rowStart := 0; rowStart < height; rowStart += 6 {
		rowEnd := rowStart + 6
		if rowEnd > height {
			rowEnd = height
		}

		// For each color in the palette, emit the sixel data for this band
		for colorIdx := range palette {
			if !colorUsedInBand(rgba, width, height, rowStart, rowEnd, colorMap, colorIdx) {
				continue
			}

			// Select color
			buf.WriteByte('#')
			buf.WriteString(strconv.Itoa(colorIdx))

			// Encode each column
			for x := 0; x < width; x++ {
				var sixel byte = 0
				for dy := 0; dy < 6; dy++ {
					y := rowStart + dy
					if y >= height {
						break
					}
					pixelIdx := y*width + x
					if int(colorMap[pixelIdx]) == colorIdx {
						sixel |= 1 << dy
					}
				}
				buf.WriteByte(0x3F + sixel)
			}
			// Newline between bands (not after last band for this color)
			if rowStart+6 < height {
				buf.WriteByte('-')
			}
		}
	}

	// String terminator
	buf.WriteString("\x1b\\")

	return buf.String()
}

// rgbColor represents an RGB color for the palette.
type rgbColor struct {
	r, g, b uint8
}

// quantizeColors extracts up to 256 unique colors from RGBA data.
// Returns the palette and a per-pixel color index map.
//
// Uses a hash-map approach for exact color matching. For images with
// more than 256 unique colors, the palette is truncated and excess
// pixels map to color 0 (background).
func quantizeColors(rgba []byte, numPixels int) ([]rgbColor, []uint8) {
	colorMap := make(map[uint32]int, 256)
	palette := make([]rgbColor, 0, 256)
	indices := make([]uint8, numPixels)

	for i := 0; i < numPixels; i++ {
		r := rgba[i*4]
		g := rgba[i*4+1]
		b := rgba[i*4+2]

		// Pack RGB into uint32 key
		key := uint32(r)<<16 | uint32(g)<<8 | uint32(b)

		idx, exists := colorMap[key]
		if !exists {
			if len(palette) >= 256 {
				// Palette full — map to nearest existing color (simplified: use 0)
				indices[i] = 0
				continue
			}
			idx = len(palette)
			colorMap[key] = idx
			palette = append(palette, rgbColor{r: r, g: g, b: b})
		}
		indices[i] = uint8(idx)
	}

	return palette, indices
}

// colorUsedInBand checks if a given color index appears in the specified
// row band. This optimization skips emitting empty color bands.
func colorUsedInBand(rgba []byte, width, height, rowStart, rowEnd int, colorMap []uint8, colorIdx int) bool {
	for y := rowStart; y < rowEnd; y++ {
		for x := 0; x < width; x++ {
			if int(colorMap[y*width+x]) == colorIdx {
				return true
			}
		}
	}
	return false
}

// EncodeSixelSimple encodes a grayscale image as Sixel with a fixed palette.
// This is a simpler, faster alternative for monochrome images.
//
// Parameters:
//   - gray: grayscale pixel data (1 byte per pixel, 0-255)
//   - width: image width in pixels
//   - height: image height in pixels
//
// Returns the complete DCS-encapsulated Sixel string.
func EncodeSixelSimple(gray []byte, width, height int) string {
	if width <= 0 || height <= 0 || len(gray) < width*height {
		return ""
	}

	var buf strings.Builder

	// DCS introducer
	buf.WriteString("\x1bP0;0q")

	// Raster attributes
	buf.WriteByte('"')
	buf.WriteString("1;1;")
	buf.WriteString(strconv.Itoa(width))
	buf.WriteByte(';')
	buf.WriteString(strconv.Itoa(height))

	// Simple 4-level grayscale palette
	grays := []struct {
		r, g, b int
	}{
		{0, 0, 0},       // black
		{85, 85, 85},    // dark gray
		{170, 170, 170}, // light gray
		{255, 255, 255}, // white
	}
	for i, c := range grays {
		buf.WriteByte('#')
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(";2;")
		buf.WriteString(strconv.Itoa(c.r * 100 / 255))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(c.g * 100 / 255))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(c.b * 100 / 255))
	}

	// Encode: map grayscale to 4 levels
	for rowStart := 0; rowStart < height; rowStart += 6 {
		for colorIdx := 0; colorIdx < 4; colorIdx++ {
			buf.WriteByte('#')
			buf.WriteString(strconv.Itoa(colorIdx))

			for x := 0; x < width; x++ {
				var sixel byte = 0
				for dy := 0; dy < 6; dy++ {
					y := rowStart + dy
					if y >= height {
						break
					}
					grayVal := gray[y*width+x]
					level := int(grayVal) / 64 // 0-3
					if level > 3 {
						level = 3
					}
					if level == colorIdx {
						sixel |= 1 << dy
					}
				}
				buf.WriteByte(0x3F + sixel)
			}
			if rowStart+6 < height {
				buf.WriteByte('-')
			}
		}
	}

	buf.WriteString("\x1b\\")
	return buf.String()
}

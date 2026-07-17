package ui

import (
	"image/color"
	"log"
	"strconv"
	"strings"
)

// RGBFromHex parses a "RRGGBB" hex color string (with or without a leading
// '#') into a color.Color.
//
// By default the color is fully opaque. An optional opacity percentage
// (0-100) can be passed as the second argument: 0 means fully transparent
// and 100 means fully opaque.
func RGBFromHex(hex string, opacityOpt ...uint8) color.Color {
	hex = strings.TrimPrefix(hex, "#")

	opacity := uint8(100)
	if len(opacityOpt) > 0 {
		opacity = opacityOpt[0]
		if opacity > 100 {
			log.Fatalf("rgbFromHex: invalid opacity %d, must be 0-100", opacity)
		}
	}

	if len(hex) != 6 {
		log.Fatalf("rgbFromHex: invalid hex length %d", len(hex))
	}

	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		log.Fatalf("rgbFromHex: parse hex %q: %s", hex, err)
	}

	// Convert 0-100 opacity (100 = fully opaque) to a 0-255 alpha value.
	alpha := uint8(int(opacity) * 0xff / 100)

	return color.NRGBA{
		R: uint8(value >> 16),
		G: uint8(value >> 8),
		B: uint8(value),
		A: alpha,
	}
}

// LightenRGB returns a copy of c with each RGB channel increased by
// amount, clamped to the valid 0-255 range.
func LightenRGB(c color.Color, amount int) color.Color {
	rgba := color.NRGBAModel.Convert(c).(color.NRGBA) //nolint:forcetypeassert

	change := func(val uint8) uint8 {
		res := int(val) + amount
		if res > 255 {
			return 255
		}
		if res < 0 {
			return 0
		}
		return uint8(res)
	}

	rgba.R = change(rgba.R)
	rgba.G = change(rgba.G)
	rgba.B = change(rgba.B)

	return rgba
}

// DarkenRGB returns a copy of c with each RGB channel decreased by
// amount, clamped to the valid 0-255 range.
func DarkenRGB(c color.Color, amount int) color.Color {
	return LightenRGB(c, -amount)
}

// LerpColor linearly interpolates between colors a and b by t, where
// t=0 returns a and t=1 returns b.
func LerpColor(a, b color.Color, t float64) color.NRGBA {
	c1 := color.NRGBAModel.Convert(a).(color.NRGBA) //nolint:forcetypeassert
	c2 := color.NRGBAModel.Convert(b).(color.NRGBA) //nolint:forcetypeassert

	lerp := func(x, y uint8, t float64) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}

	return color.NRGBA{
		R: lerp(c1.R, c2.R, t),
		G: lerp(c1.G, c2.G, t),
		B: lerp(c1.B, c2.B, t),
		A: lerp(c1.A, c2.A, t),
	}
}

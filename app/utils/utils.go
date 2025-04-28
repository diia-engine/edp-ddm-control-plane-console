package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"

	bluemonday "github.com/microcosm-cc/bluemonday"
)

func GetSvgSanitizerPolicy() *bluemonday.Policy {
	// Create a policy that allows all SVG elements but removes scripts and events
	p := bluemonday.NewPolicy()
	// Allow all SVG elements
	p.AllowElements("svg", "path", "circle", "ellipse", "line", "polygon", "polyline",
		"rect", "g", "text", "tspan", "use", "defs", "clipPath", "mask",
		"pattern", "image", "title", "desc")
	// Allow all attributes except JavaScript events (on*)
	p.AllowStyling()
	p.AllowStandardAttributes()
	p.AllowDataAttributes()
	p.AllowAttrs("xmlns", "xmlns:xlink", "viewBox", "width", "height", "fill", "stroke",
		"stroke-width", "d", "points", "x", "y", "x1", "y1", "x2", "y2", "cx", "cy",
		"r", "rx", "ry", "transform", "font-family", "font-size", "font-weight", "text-anchor").Globally()
	// allow only internal links for xlink:href
	p.AllowAttrs("xlink:href").Matching(regexp.MustCompile(`^#[\w\-]+$`)).Globally()
	return p
}

// SvgSanitizePolicy is an exported policy for SVG sanitization
var SvgSanitizePolicy *bluemonday.Policy

// init is automatically called when the package is imported
func init() {
	SvgSanitizePolicy = GetSvgSanitizerPolicy()
}

func SanitizeSvg(svgBytes string) string {
	sanitizedSVG := SvgSanitizePolicy.Sanitize(svgBytes)
	return sanitizedSVG
}

func SanitizeBase64SVG(base64SVG string) (string, error) {
	// Step 1: Decode the base64 string to get the SVG content
	svgBytes, err := base64.StdEncoding.DecodeString(base64SVG)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	svgString := string(svgBytes)

	// Step 2: Sanitize the SVG content
	sanitizedSVG := SanitizeSvg(svgString)

	// Step 3: Re-encode the sanitized SVG to base64
	sanitizedBase64 := base64.StdEncoding.EncodeToString([]byte(sanitizedSVG))

	return sanitizedBase64, nil
}

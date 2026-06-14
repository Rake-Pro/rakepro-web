// Package web embeds the static assets and HTML templates into the binary so
// the server ships as a single self-contained artifact with no runtime file
// dependencies.
package web

import "embed"

// Static holds everything under static/ (css, js, img).
//
//go:embed static
var Static embed.FS

// Templates holds the parsed HTML templates under templates/.
//
//go:embed templates
var Templates embed.FS

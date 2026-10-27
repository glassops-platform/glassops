// Package schemas embeds the published JSON Schemas so the binary validates against the
// same files the repository publishes.
package schemas

import _ "embed"

//go:embed config.json
var Config []byte

//go:embed result.json
var Result []byte

package templates

import "embed"

//go:embed *.html
var HTMLTemplates embed.FS

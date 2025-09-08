// Package docs provides embedded Swagger UI files for API documentation.
package docs

import "embed"

// SwaggerUI holds the embedded Swagger UI files.
//
//go:embed swagger-ui
var SwaggerUI embed.FS

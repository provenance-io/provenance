// Package docs provides embedded Swagger UI files for API documentation.
package docs

import "embed"

//go:embed swagger-ui
// SwaggerUI holds the embedded Swagger UI files.
var SwaggerUI embed.FS

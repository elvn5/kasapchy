// Package migrations embeds the users service SQL migrations (goose format).
// Add a new file as NNNNN_description.sql; migrations are applied on service start.
package migrations

import "embed"

// FS contains every *.sql migration of the service.
//
//go:embed *.sql
var FS embed.FS

package migrations

import "embed"

// Files contains the SQL migration assets embedded at build time.
//
//go:embed *.sql
var Files embed.FS

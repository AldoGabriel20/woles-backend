// Package migration embeds all SQL migration files so the binary carries them
// without needing the files on disk at runtime.
package migration

import "embed"

// FS holds all *.sql migration files. Pass it to goose.SetBaseFS when running
// migrations programmatically:
//
//	goose.SetBaseFS(migration.FS)
//	goose.Up(db, ".")
//
//go:embed *.sql
var FS embed.FS

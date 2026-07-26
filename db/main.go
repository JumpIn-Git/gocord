package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS

// we need to export it here since go:embed doesnt support ../ paths

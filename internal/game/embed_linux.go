//go:build linux

package game

import "embed"

//go:embed Aurora/Build/Aurora.so
var newEmbeddedFiles embed.FS

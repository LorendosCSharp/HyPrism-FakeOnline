//go:build windows

package game

import "embed"

//go:embed Aurora/Build/Aurora.dll
var newEmbeddedFiles embed.FS

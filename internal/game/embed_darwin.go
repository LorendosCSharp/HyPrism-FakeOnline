//go:build darwin

package game

import "embed"

// needs to be build , and I have no clue how to do so
// Till I will use the linux embed

//go:embed Aurora/Build/Aurora.so
var newEmbeddedFiles embed.FS

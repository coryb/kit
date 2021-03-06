package publish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hinshun/kit/config"
	"github.com/hinshun/kitapi/kit"
	shell "github.com/ipfs/go-ipfs-api"
)

func Publish(sh *shell.Shell, pluginPaths []string) (digest string, err error) {
	manifest := config.Manifest{
		Type: config.CommandManifest,
	}

	var nativePluginPath string
	for _, pluginPath := range pluginPaths {
		parts := strings.Split(pluginPath, "-")
		if len(parts) != 3 {
			return digest, fmt.Errorf("expected plugin path to be name-GOOS-GOARCH")
		}

		if parts[1] == runtime.GOOS && parts[2] == runtime.GOARCH {
			nativePluginPath = pluginPath
		}

		f, err := os.Open(pluginPath)
		if err != nil {
			return digest, err
		}

		digest, err = sh.Add(f)
		if err != nil {
			return digest, err
		}

		manifest.Platforms = append(manifest.Platforms, config.Platform{
			OS:           parts[1],
			Architecture: parts[2],
			Digest:       digest,
		})
	}

	if nativePluginPath == "" {
		return digest, fmt.Errorf("expected one of plugin path to be native")
	}

	constructor, err := kit.OpenConstructor(nativePluginPath)
	if err != nil {
		return
	}

	cmd, err := constructor()
	if err != nil {
		return digest, err
	}

	manifest.Usage = cmd.Usage()

	for _, arg := range cmd.Args() {
		manifest.Args = append(manifest.Args, config.Arg{
			Type:  arg.Type(),
			Usage: arg.Usage(),
		})
	}

	for _, flag := range cmd.Flags() {
		manifest.Flags = append(manifest.Flags, config.Flag{
			Name:  flag.Name(),
			Type:  flag.Type(),
			Usage: flag.Usage(),
		})
	}

	content, err := json.MarshalIndent(&manifest, "", "    ")
	if err != nil {
		return digest, err
	}

	return sh.Add(bytes.NewReader(content))
}

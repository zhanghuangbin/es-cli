package cmd

import (
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
	"gopkg.in/yaml.v3"
)

func formatYAML(result *types.Result, w io.Writer) error {
	data := rowsToMaps(result)
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}

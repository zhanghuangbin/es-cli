package cmd

import (
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
	"gopkg.in/yaml.v3"
)

type yamlFormatter struct{}

func (f *yamlFormatter) Format(result *types.Result, w io.Writer) error {
	data := rowsToMaps(result)
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}

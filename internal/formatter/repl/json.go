package repl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

type JsonFormatter struct{}

func (f *JsonFormatter) Format(result *types.Result, w io.Writer) error {
	if result.Meta.Endpoint != "" {
		fmt.Fprintf(w, "接口: %s\n", result.Meta.Endpoint)
	}

	if result.Source != "" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(result.Source), "", "  "); err != nil {
			fmt.Fprint(w, result.Source)
		} else {
			buf.WriteTo(w)
		}
		fmt.Fprintln(w)
	}

	return nil
}

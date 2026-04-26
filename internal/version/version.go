package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func Print(w io.Writer) {
	fmt.Fprintf(w, "es-cli 版本: %s\n", Version)
	fmt.Fprintf(w, "Git 提交: %s\n", GitCommit)
	fmt.Fprintf(w, "构建日期: %s\n", BuildDate)
	fmt.Fprintf(w, "Go 版本: %s\n", runtime.Version())
}

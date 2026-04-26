package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cmdfmt "github.com/zhanghuangbin/es-cli/internal/formatter/cmd"
	"github.com/zhanghuangbin/es-cli/internal/handler"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

var (
	execSQL      string
	execFormat   string
	execJSONPath string
	execTemplate string
	execFields   []string
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "执行 SQL 并输出结果",
	Long:  "执行单条 SQL 语句并以指定格式输出结果，适用于脚本化和管道化使用。",
	RunE: func(cmd *cobra.Command, args []string) error {
		if execSQL == "" {
			return fmt.Errorf("必须通过 -c/--command 指定 SQL 语句")
		}

		if execJSONPath != "" && execFormat != "json" {
			return fmt.Errorf("--jsonpath 仅在 json 格式下有效")
		}
		if execTemplate != "" && execFormat != "go-template" {
			return fmt.Errorf("--template 仅在 go-template 格式下有效")
		}
		if len(execFields) > 0 && execFormat != "csv" {
			return fmt.Errorf("--field 仅在 csv 格式下有效")
		}

		client, err := newESClient()
		if err != nil {
			return err
		}

		sqlType := types.DetectSQLType(execSQL)

		handlers := map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: handler.NewQueryHandler(client),
			types.SQLTypeInsert: handler.NewInsertHandler(client),
			types.SQLTypeUpdate: handler.NewUpdateHandler(client),
			types.SQLTypeDelete: handler.NewDeleteHandler(client),
			types.SQLTypeCreate: handler.NewCreateHandler(client),
			types.SQLTypeDrop:   handler.NewDropHandler(client),
			types.SQLTypeAlter:  handler.NewAlterHandler(client),
		}

		h, ok := handlers[sqlType]
		if !ok {
			return fmt.Errorf("不支持的 SQL 类型: %s", sqlType)
		}

		result, err := h.Execute(context.Background(), execSQL)
		if err != nil {
			return err
		}

		result.Meta.Type = sqlType

		fmtr, err := cmdfmt.NewFormatter(cmdfmt.Options{
			Format:   execFormat,
			JSONPath: execJSONPath,
			Template: execTemplate,
			Fields:   execFields,
		})
		if err != nil {
			return err
		}

		return fmtr.Format(result, os.Stdout)
	},
}

func init() {
	execCmd.Flags().StringVarP(&execSQL, "command", "c", "", "要执行的 SQL 语句（必填）")
	execCmd.Flags().StringVarP(&execFormat, "format", "f", "json", "输出格式（json, csv, yaml, go-template）")
	execCmd.Flags().StringVar(&execJSONPath, "jsonpath", "", "JSONPath 表达式（仅 json 格式有效）")
	execCmd.Flags().StringVar(&execTemplate, "template", "", "Go 模板字符串（仅 go-template 格式有效）")
	execCmd.Flags().StringSliceVar(&execFields, "field", nil, "输出的列名（仅 csv 格式有效，可多次指定）")
	rootCmd.AddCommand(execCmd)
}

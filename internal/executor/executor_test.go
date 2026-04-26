package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/zhanghuangbin/es-cli/internal/handler"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

// mockHandler 用于测试的 Handler 实现
type mockHandler struct {
	result *types.Result
	err    error
}

func (m *mockHandler) Execute(ctx context.Context, sql string) (*types.Result, error) {
	return m.result, m.err
}

// 确保 mockHandler 实现了 handler.Handler 接口
var _ handler.Handler = (*mockHandler)(nil)

// mockFormatter 用于测试的 Formatter 实现
type mockFormatter struct {
	lastResult *types.Result
	err        error
}

func (m *mockFormatter) Format(result *types.Result, w io.Writer) error {
	m.lastResult = result
	if m.err != nil {
		return m.err
	}
	fmt.Fprint(w, "formatted output")
	return nil
}

func TestExecutorExecute(t *testing.T) {
	result := &types.Result{
		Meta:    types.Meta{Status: 200},
		Columns: []string{"name"},
		Rows:    [][]any{{"test"}},
	}

	h := &mockHandler{result: result}
	fmtr := &mockFormatter{}
	var buf bytes.Buffer

	exec := &Executor{
		formatter: fmtr,
		output:    &buf,
		handlers: map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: h,
		},
	}

	err := exec.Execute("SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fmtr.lastResult == nil {
		t.Fatal("formatter was not called")
	}

	if fmtr.lastResult.Meta.Type != types.SQLTypeSelect {
		t.Errorf("result.Meta.Type = %v, want SQLTypeSelect", fmtr.lastResult.Meta.Type)
	}

	if buf.String() != "formatted output" {
		t.Errorf("output = %q, want 'formatted output'", buf.String())
	}
}

func TestExecutorExecuteHandlerError(t *testing.T) {
	h := &mockHandler{err: fmt.Errorf("handler error")}
	fmtr := &mockFormatter{}
	var buf bytes.Buffer

	exec := &Executor{
		formatter: fmtr,
		output:    &buf,
		handlers: map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: h,
		},
	}

	err := exec.Execute("SELECT * FROM users")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "handler error" {
		t.Errorf("error = %q, want 'handler error'", err.Error())
	}
}

func TestExecutorExecuteFormatterError(t *testing.T) {
	result := &types.Result{
		Meta:    types.Meta{Status: 200},
		Columns: []string{"name"},
		Rows:    [][]any{{"test"}},
	}

	h := &mockHandler{result: result}
	fmtr := &mockFormatter{err: fmt.Errorf("format error")}
	var buf bytes.Buffer

	exec := &Executor{
		formatter: fmtr,
		output:    &buf,
		handlers: map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: h,
		},
	}

	err := exec.Execute("SELECT * FROM users")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "format error" {
		t.Errorf("error = %q, want 'format error'", err.Error())
	}
}

func TestExecutorExecuteUnsupportedType(t *testing.T) {
	fmtr := &mockFormatter{}
	var buf bytes.Buffer

	exec := &Executor{
		formatter: fmtr,
		output:    &buf,
		handlers:  map[types.SQLType]handler.Handler{},
	}

	err := exec.Execute("SELECT * FROM users")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestExecutorSetFormatter(t *testing.T) {
	fmtr1 := &mockFormatter{}
	fmtr2 := &mockFormatter{}
	var buf bytes.Buffer

	result := &types.Result{
		Meta:    types.Meta{Status: 200},
		Columns: []string{"a"},
		Rows:    [][]any{{"b"}},
	}
	h := &mockHandler{result: result}

	exec := &Executor{
		formatter: fmtr1,
		output:    &buf,
		handlers: map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: h,
		},
	}

	exec.SetFormatter(fmtr2)
	exec.Execute("SELECT 1")

	if fmtr2.lastResult == nil {
		t.Error("new formatter should have been called")
	}
	if fmtr1.lastResult != nil {
		t.Error("old formatter should not have been called")
	}
}

func TestExecutorDispatchesByType(t *testing.T) {
	selectResult := &types.Result{Meta: types.Meta{Status: 200, Message: "select"}, Columns: []string{"a"}, Rows: [][]any{{"b"}}}
	insertResult := &types.Result{Meta: types.Meta{Status: 200, Message: "insert"}, Columns: []string{"a"}, Rows: [][]any{{"b"}}}

	selectHandler := &mockHandler{result: selectResult}
	insertHandler := &mockHandler{result: insertResult}
	fmtr := &mockFormatter{}
	var buf bytes.Buffer

	exec := &Executor{
		formatter: fmtr,
		output:    &buf,
		handlers: map[types.SQLType]handler.Handler{
			types.SQLTypeSelect: selectHandler,
			types.SQLTypeInsert: insertHandler,
		},
	}

	// 测试 SELECT 分发
	exec.Execute("SELECT 1")
	if fmtr.lastResult.Meta.Message != "select" {
		t.Errorf("SELECT should dispatch to select handler, got message %q", fmtr.lastResult.Meta.Message)
	}

	// 测试 INSERT 分发
	buf.Reset()
	exec.Execute("INSERT INTO t (a) VALUES (1)")
	if fmtr.lastResult.Meta.Message != "insert" {
		t.Errorf("INSERT should dispatch to insert handler, got message %q", fmtr.lastResult.Meta.Message)
	}
}

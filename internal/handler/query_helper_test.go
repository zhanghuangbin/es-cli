package handler

import "testing"

func TestBuildResultFromSource(t *testing.T) {
	source := map[string]any{
		"name": "张三",
		"age":  float64(25),
	}

	result := buildResultFromSource("doc123", source)

	// 验证 columns：_id 在最前面，其余按字母排序
	if len(result.Columns) != 3 {
		t.Fatalf("columns count = %d, want 3", len(result.Columns))
	}
	if result.Columns[0] != "_id" {
		t.Errorf("columns[0] = %q, want '_id'", result.Columns[0])
	}
	// age 在 name 前面（字母排序）
	if result.Columns[1] != "age" {
		t.Errorf("columns[1] = %q, want 'age'", result.Columns[1])
	}
	if result.Columns[2] != "name" {
		t.Errorf("columns[2] = %q, want 'name'", result.Columns[2])
	}

	// 验证 rows
	if len(result.Rows) != 1 {
		t.Fatalf("rows count = %d, want 1", len(result.Rows))
	}
	row := result.Rows[0]
	if len(row) != 3 {
		t.Fatalf("row length = %d, want 3", len(row))
	}
	if row[0] != "doc123" {
		t.Errorf("row[0] (_id) = %v, want 'doc123'", row[0])
	}
	if row[1] != float64(25) {
		t.Errorf("row[1] (age) = %v, want 25", row[1])
	}
	if row[2] != "张三" {
		t.Errorf("row[2] (name) = %v, want '张三'", row[2])
	}

	// 验证 meta
	if result.Meta.Status != 200 {
		t.Errorf("meta.Status = %d, want 200", result.Meta.Status)
	}
	if result.Meta.Message != "操作成功" {
		t.Errorf("meta.Message = %q, want '操作成功'", result.Meta.Message)
	}
}

func TestBuildResultFromSourceEmpty(t *testing.T) {
	source := map[string]any{}
	result := buildResultFromSource("id1", source)

	// 只有 _id 列
	if len(result.Columns) != 1 {
		t.Fatalf("columns count = %d, want 1", len(result.Columns))
	}
	if result.Columns[0] != "_id" {
		t.Errorf("columns[0] = %q, want '_id'", result.Columns[0])
	}
	if len(result.Rows) != 1 {
		t.Fatalf("rows count = %d, want 1", len(result.Rows))
	}
	if result.Rows[0][0] != "id1" {
		t.Errorf("row[0] = %v, want 'id1'", result.Rows[0][0])
	}
}

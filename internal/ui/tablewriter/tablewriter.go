package tablewriter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Column 表格列定义
// 定义表格中每一列的属性
type Column struct {
	Name         string // 列名
	SeparateLine bool   // 是否单独一行显示
	RightAlign   bool   // 是否右对齐
}

// columnCfg 列配置
// 用于配置列的显示选项
type columnCfg struct {
	rightAlign bool // 右对齐标志
}

// ColumnOption 列选项函数类型
// 用于配置列的显示选项
type ColumnOption func(*columnCfg)

// RightAlign 返回右对齐选项
// 用于设置列内容右对齐
func RightAlign() ColumnOption {
	return func(c *columnCfg) {
		c.rightAlign = true
	}
}

// TableWriter 表格写入器
// 用于格式化输出表格数据
type TableWriter struct {
	cols []Column            // 列定义
	rows []map[string]string // 行数据
}

// Col 创建普通列
// 根据给定的列名和选项创建列定义
func Col(name string, opts ...ColumnOption) Column {
	cfg := &columnCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return Column{
		Name:         name,
		SeparateLine: false,
		RightAlign:   cfg.rightAlign,
	}
}

// NewLineCol 创建单独行列
// 创建一个在单独行显示的列
func NewLineCol(name string) Column {
	return Column{
		Name:         name,
		SeparateLine: true,
	}
}

// New 创建新的表格写入器
// 使用给定的列定义初始化表格写入器
func New(cols ...Column) *TableWriter {
	return &TableWriter{
		cols: cols,
	}
}

// Write 写入一行数据
// 将 map 数据转换为字符串并添加到表格行中
func (w *TableWriter) Write(r map[string]interface{}) {
	row := make(map[string]string, len(r))
	for k, v := range r {
		row[k] = fmt.Sprint(v)
	}
	w.rows = append(w.rows, row)
}

// Flush 将表格数据输出到指定的写入器
// 使用 tabwriter 格式化输出表格，支持普通列和单独行列
func (w *TableWriter) Flush(out io.Writer) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	// 收集非单独行的列作为表头
	headerCols := make([]string, 0, len(w.cols))
	for _, col := range w.cols {
		if col.SeparateLine {
			continue
		}
		headerCols = append(headerCols, col.Name)
	}
	// 输出表头
	if len(headerCols) > 0 {
		if _, err := fmt.Fprintln(tw, strings.Join(headerCols, "\t")); err != nil {
			return err
		}
	}

	// 输出每一行数据
	for _, row := range w.rows {
		// 输出普通列
		fields := make([]string, 0, len(headerCols))
		for _, col := range w.cols {
			if col.SeparateLine {
				continue
			}
			fields = append(fields, row[col.Name])
		}
		if len(fields) > 0 {
			if _, err := fmt.Fprintln(tw, strings.Join(fields, "\t")); err != nil {
				return err
			}
		}

		// 输出单独行列
		for _, col := range w.cols {
			if !col.SeparateLine {
				continue
			}
			if val := row[col.Name]; val != "" {
				if _, err := fmt.Fprintf(tw, "  %s:\t%s\n", col.Name, val); err != nil {
					return err
				}
			}
		}
	}

	return tw.Flush()
}

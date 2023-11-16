package data_test

import (
	"context"
	"testing"

	"permission/components"
	"permission/helpers"

	"permission/pkg/golib/v2/hbase"
	"permission/pkg/golib/v2/zlog"
)

func TestHBase_GetTable(t *testing.T) {
	var err error
	var tables []hbase.Text
	efunc := func(c *hbase.HbaseClient) error {
		tables, err = c.GetTableNames(context.Background())
		return err
	}

	err = helpers.HBaseDemo.Exec(ctx, efunc)
	if err != nil {
		t.Error("[TestHBase_GetTable]  error: ", err.Error())
		return
	}

	for _, v := range tables {
		zlog.Debug(ctx, string(v))
	}

}
func TestHBase_Query(t *testing.T) {
	var err error
	tbl := hbase.Text("sm:homework_user_message")

	// 查询表
	var results []*hbase.TRowResult_
	efunc := func(c *hbase.HbaseClient) error {
		results, err = c.GetRow(
			context.Background(),
			tbl,
			hbase.Text("c"),
			nil,
		)
		if err != nil {
			return components.ErrorHbaseQuery.Wrap(err)
		}
		return err
	}

	err = helpers.HBaseDemo.Exec(ctx, efunc)
	if err != nil {
		t.Error("query exec error: ", err.Error())
		return
	}

	for _, vs := range results {
		for k, v := range vs.Columns {
			zlog.Debug(ctx, "k= ", k, " v=", string(v.Value))
		}
	}
}

func TestHBase_GetColumnDescriptors(t *testing.T) {
	var err error
	tbl := hbase.Text("sm:homework_user_message")

	var colDes map[string]*hbase.ColumnDescriptor
	efunc := func(c *hbase.HbaseClient) error {
		colDes, err = c.GetColumnDescriptors(context.Background(), tbl)
		if err != nil {
			zlog.Error(ctx, "GetColumnDescriptors error: ", err.Error())
		}

		return err
	}
	err = helpers.HBaseDemo.Exec(ctx, efunc)
	if err != nil {
		t.Error("[TestHBase_GetColumnDescriptors] error: ", err.Error())
		return
	}

	for k, v := range colDes {
		zlog.Error(ctx, "k= ", k, " v=", string(v.Name))
	}
}

// 如果一次需要多个操作，可以采用如下方式，拿到一个链接后，执行完所有操作再归还链接
func TestHBase_Do(t *testing.T) {
	var err error
	var tables []hbase.Text
	efunc := func(c *hbase.HbaseClient) error {
		tables, err = c.GetTableNames(context.Background())
		return err
	}

	conn, err := helpers.HBaseDemo.GetConn(ctx)
	if err != nil {
		println("get conn error: ", err.Error())
		return
	}
	defer helpers.HBaseDemo.Release(conn)

	err = conn.Do(ctx, efunc)
	if err != nil {
		t.Error("[TestHBase_Do] error: ", err.Error())
		return
	}

	for _, v := range tables {
		zlog.Debug(ctx, string(v))
	}

	tbl := hbase.Text("user")
	var colDes map[string]*hbase.ColumnDescriptor
	efunc = func(c *hbase.HbaseClient) error {
		colDes, err = c.GetColumnDescriptors(context.Background(), tbl)
		if err != nil {
			zlog.Error(ctx, "GetColumnDescriptors error: ", err.Error())
		}

		return err
	}
	err = conn.Do(ctx, efunc)
	if err != nil {
		t.Error("[TestHBase_Do] conn Do error: ", err.Error())
		return
	}

	for k, v := range colDes {
		zlog.Error(ctx, "k= ", k, " v=", string(v.Name))
		println(ctx, "k= ", k, " v=", string(v.Name))
	}

	// 查询表
	var results []*hbase.TRowResult_
	efunc = func(c *hbase.HbaseClient) error {
		results, err = c.GetRow(
			context.Background(),
			tbl,
			hbase.Text("c"),
			nil,
		)
		if err != nil {
			return components.ErrorHbaseQuery.Wrap(err)
		}
		return err
	}
	err = conn.Do(ctx, efunc)
	for _, vs := range results {
		for k, v := range vs.Columns {
			zlog.Debug(ctx, "k= ", k, " v=", string(v.Value))
		}
	}
}

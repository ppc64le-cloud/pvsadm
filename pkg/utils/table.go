package utils

import (
	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
	"k8s.io/klog/v2"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type Table struct {
	*tablewriter.Table
	doOnce sync.Once
}

func NewTable() *Table {
	t := &Table{}
	t.Table = tablewriter.NewWriter(os.Stdout)
	return t
}

func (t *Table) SetHeader(keys []string) {
	t.doOnce.Do(func() {
		t.Table.SetHeader(keys)
	})
}

func (t *Table) Render(rows interface{}, fields []string) {
	switch reflect.TypeOf(rows).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(rows)
		for i := 0; i < s.Len(); i++ {
			var headers []string
			val := s.Index(i).Elem()
			var row []string
			for i := 0; i < val.NumField(); i++ {
				if f := strings.ToLower(val.Type().Field(i).Name); f == "href" || f == "specifications" {
					continue
				}
				headers = append(headers, val.Type().Field(i).Name)
				content := reflect.Indirect(reflect.ValueOf(val.Field(i).Interface()))
				row = append(row, getcontent(content))
			}
			t.SetHeader(headers)
			t.Append(row)
		}
	}
	t.Table.Render()
}

func getcontent(value reflect.Value) (strVal string) {
	if value.Kind() == reflect.Invalid {
		return ""
	}
	switch value.Interface().(type) {
	case int, int8, int16, int32, int64:
		strVal = strconv.FormatInt(value.Int(), 10)
	case bool:
		strVal = strconv.FormatBool(value.Bool())
	case float32, float64:
		strVal = strconv.FormatFloat(value.Float(), 'f', -1, 64)
	case string:
		strVal = value.String()
	case strfmt.DateTime:
		t := value.Interface().(strfmt.DateTime)
		strVal = t.String()
	default:
		switch value.Kind() {
		case reflect.Slice:
			var st []string
			for i := 0; i < value.Len(); i++ {
				st = append(st, getcontent(value.Index(i)))
			}
			strVal = strings.Join(st, ",")
		default:
			klog.Infof("I'm here at default type, unable to handle: %s", value.Kind())
			strVal = value.String()
		}
	}
	return
}

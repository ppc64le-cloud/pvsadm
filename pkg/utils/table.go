package utils

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
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

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func (t *Table) Render(rows interface{}, exclude []string) {
	noData := true
	switch reflect.TypeOf(rows).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(rows)
		for i := 0; i < s.Len(); i++ {
			noData = false
			var headers []string
			val := s.Index(i).Elem()
			var row []string
			for i := 0; i < val.NumField(); i++ {
				if f := strings.ToLower(val.Type().Field(i).Name); Contains(exclude, f) {
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
	if noData {
		fmt.Println("\n--NO DATA FOUND--")
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
			strVal = fmt.Sprintf("%+v", value)
		}
	}
	return
}

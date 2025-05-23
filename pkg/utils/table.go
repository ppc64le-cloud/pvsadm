// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
	"k8s.io/klog/v2"
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
		t.Table.Header(keys)
	})
}

func (t *Table) Render(rows interface{}, exclude []string) {
	noData := true
	switch reflect.TypeOf(rows).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(rows)
		for i := 0; i < s.Len(); i++ {
			noData = false
			var headers, row []string
			val := s.Index(i).Elem()
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
		klog.Info("No data found to display")
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

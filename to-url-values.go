package gobinance

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

func toURLValues(i interface{}) (url.Values, error) {
	out := url.Values{}

	iType := reflect.TypeOf(i)
	if iType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("toURLValues can only be called on structs")
	}

	iVal := reflect.ValueOf(i)
	omitEmpty := false

	for f := 0; f < iType.NumField(); f++ {
		if !iVal.Field(f).CanInterface() {
			// private
			continue
		}
		name := iType.Field(f).Name
		tagStr := iType.Field(f).Tag.Get("param")
		if tagStr != "" {
			tag := strings.Split(tagStr, ",")
			if len(tag) > 0 {
				name = tag[0]
			}
			if name == "-" {
				continue
			}
			for _, v := range tag[1:] {
				if v == "omitempty" {
					omitEmpty = true
				}
			}
		}

		if iVal.Field(f).IsZero() && omitEmpty {
			continue
		}
		out.Add(name, fmt.Sprint(iVal.Field(f).Interface()))
	}
	return out, nil
}

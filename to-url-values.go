package gobinance

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// toURLValues converts a structure into url.Values
// It respects the `param` tag, which is a comma separated list of directives.
//
// The first item in the tag is the key to be used for that field in the output url.Values.
// If that value is -, then the field is always omitted.
// If `omitempty` is provided in any of the directives after the first (i.e. the name), then
// the field will not be in the output when the value of that field is the Zero value of its type.
//
// The `emptyvalue` tag may be used to specify the value to be used in the output when the field is empty.
func toURLValues(i interface{}) (url.Values, error) {
	out := url.Values{}

	iType := reflect.TypeOf(i)
	if iType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("toURLValues can only be called on structs")
	}

	iVal := reflect.ValueOf(i)

	for f := 0; f < iType.NumField(); f++ {
		omitEmpty := false
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
				switch v {
				case "omitempty":
					omitEmpty = true
				}
			}
		}

		stringValue := fmt.Sprint(iVal.Field(f).Interface())
		if iVal.Field(f).IsZero() {
			if omitEmpty {
			continue
			}
			if v := iType.Field(f).Tag.Get("emptyvalue"); v != "" {
				stringValue = v
			}
		}
		out.Add(name, stringValue)
	}
	return out, nil
}

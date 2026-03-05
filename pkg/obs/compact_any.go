package obs

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

// CompactAny converts arbitrary values into a JSON-friendly shape (maps/slices/scalars),
// while compacting binary-like fields:
// - []byte -> hex string, shortened to first4...last4 bytes when long
// - hex strings -> shortened to first4...last4 bytes (8 hex chars each side) when long
//
// It is intended for observability payloads where users want "all fields",
// but very long/binary values should be readable.
func CompactAny(v any) any {
	seen := map[uintptr]bool{}
	return compactAny(reflect.ValueOf(v), 0, seen)
}

const maxCompactDepth = 8

func compactAny(rv reflect.Value, depth int, seen map[uintptr]bool) any {
	if !rv.IsValid() {
		return nil
	}
	if depth >= maxCompactDepth {
		return fmt.Sprint(rv.Interface())
	}

	for rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		ptr := rv.Pointer()
		if ptr != 0 {
			if seen[ptr] {
				return "<cycle>"
			}
			seen[ptr] = true
			defer delete(seen, ptr)
		}
		return compactAny(rv.Elem(), depth+1, seen)
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.String:
		return compactHexString(rv.String())
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return compactBytes(rv.Bytes())
		}
		out := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, compactAny(rv.Index(i), depth+1, seen))
		}
		return out
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(b), rv)
			return compactBytes(b)
		}
		out := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, compactAny(rv.Index(i), depth+1, seen))
		}
		return out
	case reflect.Map:
		if rv.IsNil() {
			return nil
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			key := ""
			if k.IsValid() && k.Kind() == reflect.String {
				key = k.String()
			} else {
				key = fmt.Sprint(k.Interface())
			}
			out[key] = compactAny(iter.Value(), depth+1, seen)
		}
		return out
	case reflect.Struct:
		out := make(map[string]any, rv.NumField())
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			f := rt.Field(i)
			if !f.IsExported() {
				continue
			}
			name := f.Name
			if tag, ok := f.Tag.Lookup("json"); ok {
				tagName := strings.TrimSpace(strings.Split(tag, ",")[0])
				if tagName == "-" {
					continue
				}
				if tagName != "" {
					name = tagName
				}
			}
			out[name] = compactAny(rv.Field(i), depth+1, seen)
		}
		return out
	default:
		return fmt.Sprint(rv.Interface())
	}
}

func compactBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if len(b) <= 8 {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b[:4]) + "..." + hex.EncodeToString(b[len(b)-4:])
}

func compactHexString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 16 {
		return s
	}
	if len(s)%2 != 0 {
		return s
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return s
	}
	// first4 + last4 bytes => first8 + last8 hex chars
	return s[:8] + "..." + s[len(s)-8:]
}

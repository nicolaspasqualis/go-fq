package fq

import (
	"reflect"
	"time"
)

// compareValues compares two values
func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// fast path for common types
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case time.Time:
		if bVal, ok := b.(time.Time); ok {
			if aVal.Before(bVal) {
				return -1
			} else if aVal.After(bVal) {
				return 1
			}
			return 0
		}
	}

	// normalize & check for mixed num types
	aNum, aIsNum := toNumber(a)
	bNum, bIsNum := toNumber(b)

	if aIsNum && bIsNum {
		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
		return 0
	}

	// Types are not comparable - treat as not equal
	return -1
}

// isEqual provides improved equality checking with safety for uncomparable types
func isEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}

	aNum, aIsNum := toNumber(a)
	bNum, bIsNum := toNumber(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// use DeepEqual on uncomparable type (slice, map, etc.)
	if isUncomparable(aVal.Kind()) || isUncomparable(bVal.Kind()) {
		return reflect.DeepEqual(a, b)
	}

	if aVal.Kind() == bVal.Kind() {
		return a == b
	}

	return reflect.DeepEqual(a, b) // fall back
}

// isUncomparable checks if a type is not comparable with ==
func isUncomparable(k reflect.Kind) bool {
	switch k {
	case reflect.Slice, reflect.Map, reflect.Func:
		return true
	default:
		return false
	}
}

// toNumber converts numeric types to float64 (doesn't try to parse strings)
func toNumber(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}

	// Fast path for common types (avoids reflection)
	switch val := v.(type) {
	case int:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	}

	// Fallback with reflection for less common types
	reflectV := reflect.ValueOf(v)
	switch reflectV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(reflectV.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(reflectV.Uint()), true
	case reflect.Float32, reflect.Float64:
		return reflectV.Float(), true
	default:
		return 0, false
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	kind := val.Kind()

	return (kind == reflect.Ptr || kind == reflect.Slice ||
		kind == reflect.Map || kind == reflect.Interface ||
		kind == reflect.Chan || kind == reflect.Func) && val.IsNil()
}

package fq

import (
	"math"
	"reflect"
	"regexp"
	"strings"
)

// Eq checks for equality
func Eq(val interface{}) P {
	return func(v interface{}) bool {
		return isEqual(v, val)
	}
}

// Gt checks if a value is greater than threshold
func Gt(threshold interface{}) P {
	return func(v interface{}) bool {
		return compareValues(v, threshold) > 0
	}
}

// Lt checks if a value is less than threshold
func Lt(threshold interface{}) P {
	return func(v interface{}) bool {
		return compareValues(v, threshold) < 0
	}
}

// Gte checks if a value is greater than or equal to threshold
func Gte(threshold interface{}) P {
	return func(v interface{}) bool {
		return compareValues(v, threshold) >= 0
	}
}

// Lte checks if a value is less than or equal to threshold
func Lte(threshold interface{}) P {
	return func(v interface{}) bool {
		return compareValues(v, threshold) <= 0
	}
}

// In checks if value matches any provided values
func In(vals ...interface{}) P {
	return func(v interface{}) bool {
		for _, val := range vals {
			if reflect.DeepEqual(v, val) {
				return true
			}
		}
		return false
	}
}

// Contains checks if a string contains substring
func Contains(substr string) P {
	return func(v interface{}) bool {
		if s, ok := v.(string); ok {
			return strings.Contains(s, substr)
		}
		return false
	}
}

// HasItem checks if an array contains the item
func HasItem(item interface{}) P {
	return func(v interface{}) bool {
		switch arr := v.(type) {
		case []interface{}:
			for _, val := range arr {
				if reflect.DeepEqual(val, item) {
					return true
				}
			}
			return false
		case []string:
			if str, ok := item.(string); ok {
				for _, val := range arr {
					if val == str {
						return true
					}
				}
			}
			return false
		}

		arr := reflect.ValueOf(v)
		if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
			return false
		}

		for i := 0; i < arr.Len(); i++ {
			if reflect.DeepEqual(arr.Index(i).Interface(), item) {
				return true
			}
		}
		return false
	}
}

// GeoWithin checks if a location is within a given radius of a center point using the Haversine formula
func GeoWithin(centerLat, centerLng, radiusKm float64) P {
	return func(v interface{}) bool {
		var lat, lng float64
		var ok bool

		switch coords := v.(type) {
		case [2]float64:
			lat, lng = coords[0], coords[1]
			ok = true
		case []float64:
			if len(coords) >= 2 {
				lat, lng = coords[0], coords[1]
				ok = true
			}
		case []interface{}:
			if len(coords) >= 2 {
				lat, ok = toNumber(coords[0])
				if ok {
					lng, ok = toNumber(coords[1])
				}
			}
		default:
				rv := reflect.ValueOf(v)
			if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
				return false
			}
			if rv.Len() < 2 {
				return false
			}
			lat, ok = toNumber(rv.Index(0).Interface())
			if ok {
				lng, ok = toNumber(rv.Index(1).Interface())
			}
		}

		if !ok {
			return false
		}

		// Haversine formula for great-circle distance
		const R = 6371.0 // Earth radius in kilometers

		// Convert to radians
		lat1 := centerLat * math.Pi / 180
		lng1 := centerLng * math.Pi / 180
		lat2 := lat * math.Pi / 180
		lng2 := lng * math.Pi / 180

		dLat := lat2 - lat1
		dLng := lng2 - lng1

		a := math.Sin(dLat/2)*math.Sin(dLat/2) +
			math.Cos(lat1)*math.Cos(lat2)*
				math.Sin(dLng/2)*math.Sin(dLng/2)

		c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
		distance := R * c

		return distance <= radiusKm
	}
}

// Match checks if a string matches a pattern (string contains or regex)
func Match(pattern interface{}) P {
	return func(v interface{}) bool {
		str := ""
		switch sv := v.(type) {
		case string:
			str = sv
		default:
			str = strings.ToLower(strings.TrimSpace(reflect.ValueOf(v).String()))
		}

		switch p := pattern.(type) {
		case string:
			return strings.Contains(
				strings.ToLower(str),
				strings.ToLower(p),
			)
		case *regexp.Regexp:
			return p.MatchString(str)
		default:
			patternStr := strings.ToLower(strings.TrimSpace(reflect.ValueOf(pattern).String()))
			return strings.Contains(str, patternStr)
		}
	}
}

// ContainsAll checks if an array contains all specified items
func ContainsAll(items ...interface{}) P {
	return func(v interface{}) bool {
		arr := reflect.ValueOf(v)
		if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
			return false
		}

		for _, item := range items {
			found := false
			for i := 0; i < arr.Len(); i++ {
				if reflect.DeepEqual(arr.Index(i).Interface(), item) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}

		return true
	}
}

// ContainsAny checks if an array contains any of the specified items
func ContainsAny(items ...interface{}) P {
	return func(v interface{}) bool {
		switch arr := v.(type) {
		case []interface{}:
			for _, item := range items {
				for _, val := range arr {
					if reflect.DeepEqual(val, item) {
						return true
					}
				}
			}
			return false
		case []string:
			for _, item := range items {
				if str, ok := item.(string); ok {
					for _, val := range arr {
						if val == str {
							return true
						}
					}
				}
			}
			return false
		}

		arr := reflect.ValueOf(v)
		if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
			return false
		}
		for _, item := range items {
			for i := 0; i < arr.Len(); i++ {
				if reflect.DeepEqual(arr.Index(i).Interface(), item) {
					return true
				}
			}
		}
		return false
	}
}

// Or combines values with logical OR
func Or(vals ...Query) P {
	return func(v interface{}) bool {
		for _, val := range vals {
			if eval(val, v) {
				return true
			}
		}
		return false
	}
}

// And combines predicates with logical AND
func And(predicates ...Query) P {
	return func(v interface{}) bool {
		for _, p := range predicates {
			if !eval(p, v) {
				return false
			}
		}
		return true
	}
}

// Not negates a predicate
func Not(p Query) P {
	return func(v interface{}) bool {
		return !eval(p, v)
	}
}

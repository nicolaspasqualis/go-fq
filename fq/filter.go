package fq

import (
	"fmt"
	"reflect"
)

// Query is a generic interface for all query types
type Query interface{}

// Q represents a map-based query document
type Q map[string]interface{}

// P is a function that evaluates whether a value meets a condition
type P func(interface{}) bool

// Filter filters data based on any query type
func Filter[T any](data []T, query Query, skip int, limit int) (result []T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during filtering: %v", r)
		}
	}()

	if query == nil {
		if skip == 0 && (limit == 0 || limit >= len(data)) {
			return data, nil
		}
		return data[min(skip, len(data)):min(skip+limit, len(data))], nil
	}

	count := 0
	for _, item := range data {
		if eval(query, item) {
			if count < skip {
				count++
				continue
			}

			result = append(result, item)

			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result, err
}

// FilterC filters data based on any query type (like Filter but with channel io)
func FilterC[T any](input <-chan T, query Query, skip int, limit int) (<-chan T, <-chan error) {
	output := make(chan T)
	errCh := make(chan error)

	go func() {
		defer close(output)
		defer close(errCh)

		matched := 0
		sent := 0

		for item := range input {
			var matches bool
			func() {
				defer func() {
					if r := recover(); r != nil {
						errCh <- fmt.Errorf("panic during filter evaluation: %v", r)
						matches = false
					}
				}()
				matches = eval(query, item)
			}()

			if matches {
				matched++

				if matched <= skip {
					continue
				}
				if limit > 0 && sent >= limit {
					return
				}

				output <- item
				sent++
			}
		}
	}()

	return output, errCh
}

// eval checks if a value satisfies a query of any type
func eval(query Query, value interface{}) bool {
	switch q := query.(type) {
	case P:
		return q(value)
	case func(interface{}) bool:
		return q(value)
	case Q:
		return evalMapQuery(value, q)
	case map[string]interface{}:
		return evalMapQuery(value, q)
	case nil:
		return isNil(value)
	default:
		return isEqual(value, q)
	}
}

// evalMapQuery checks if an item matches a map-based query.
// each field in the map acts as a condition with implicit AND
func evalMapQuery(item interface{}, query Q) bool {
	for key, condition := range query {
		var value interface{}
		if key == "" {
			value = item
		} else {
			value = getField(item, key)
		}

		if !eval(condition, value) {
			return false
		}
	}

	return true
}

func getField(item interface{}, fieldName string) interface{} {
	if item == nil {
		return nil
	}

	itemVal := reflect.ValueOf(item)

	if itemVal.Kind() == reflect.Ptr {
		itemVal = itemVal.Elem()
	}

	switch itemVal.Kind() {
	case reflect.Map:
		keyVal := reflect.ValueOf(fieldName)
		valueVal := itemVal.MapIndex(keyVal)
		if !valueVal.IsValid() {
			return nil
		}
		return valueVal.Interface()

	case reflect.Struct:
		field := itemVal.FieldByName(fieldName)
		if !field.IsValid() {
			return nil
		}
		return field.Interface()

	default:
		return nil
	}
}

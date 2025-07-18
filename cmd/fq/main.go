package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/nicolaspasqualis/go-fq/fq"
)

const usage = `Usage: fq [options] <data-file> [filters...]

Options:
  -skip <number>           Skip first N results
  -limit <number>          Limit to N results
  -quiet                   Suppress error messages
  -help                    Show this help

Filters:
  field:operator:value

Operators:
  eq         Equal to
  gt         Greater than
  lt         Less than
  gte        Greater than or equal
  lte        Less than or equal
  match      Case-insensitive text match
  contains   String contains substring
  hasitem    Array contains value
  in         Value in comma-separated list
  geowithin  Geospatial within radius (lat,lon,radius)

Examples:
  fq data.jsonl "price:lt:500"
  fq data.jsonl "status:eq:active" "category:in:electronics,books"
  fq data.jsonl "location:geowithin:40.7,-74.0,10"
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(usage)
		os.Exit(1)
	}

	var skip, limit int
	var quiet, help bool

	args := os.Args[1:]
	var dataFile string
	var filters []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-skip" && i+1 < len(args):
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				skip = val
			}
			i++
		case arg == "-limit" && i+1 < len(args):
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				limit = val
			}
			i++
		case arg == "-quiet":
			quiet = true
		case arg == "-help":
			help = true
		case !strings.HasPrefix(arg, "-") && dataFile == "":
			dataFile = arg
		case !strings.HasPrefix(arg, "-"):
			filters = append(filters, arg)
		}
	}

	if help {
		fmt.Printf(usage)
		return
	}

	if dataFile == "" {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: data file required\n")
		}
		os.Exit(1)
	}

	query, err := parseFilters(filters)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error parsing filters: %v\n", err)
		}
		os.Exit(1)
	}

	dataCh, srcErrCh := fq.JSONLFileSourceStream(dataFile)
	resultCh, filterErrCh := fq.FilterC(dataCh, query, skip, limit)

	if err := process(resultCh, srcErrCh, filterErrCh, quiet); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func parseFilters(filters []string) (fq.Query, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	query := fq.Q{}
	for _, filter := range filters {
		parts := strings.SplitN(filter, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid filter format: %s", filter)
		}

		field, operator, value := parts[0], parts[1], parts[2]

		predicate, err := createPredicate(operator, value)
		if err != nil {
			return nil, err
		}

		query[field] = predicate
	}

	return query, nil
}

var operatorFuncs = map[string]interface{}{
	"eq":          fq.Eq,
	"gt":          fq.Gt,
	"lt":          fq.Lt,
	"gte":         fq.Gte,
	"lte":         fq.Lte,
	"match":       fq.Match,
	"contains":    fq.Contains,
	"hasitem":     fq.HasItem,
	"containsall": fq.ContainsAll,
	"containsany": fq.ContainsAny,
	"in":          fq.In,
	"not":         fq.Not,
	"and":         fq.And,
	"or":          fq.Or,
	"geowithin":   fq.GeoWithin,
}

func createPredicate(operator, value string) (fq.P, error) {
	fn, exists := operatorFuncs[operator]
	if !exists {
		return nil, fmt.Errorf("unknown operator: %s", operator)
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.NumIn() == 0 {
		return nil, fmt.Errorf("operator %s requires arguments", operator)
	}

	args, err := parseArgs(fnType, value)
	if err != nil {
		return nil, err
	}

	result := fnValue.Call(args)
	if len(result) != 1 {
		return nil, fmt.Errorf("unexpected return value from operator %s", operator)
	}

	predicate, ok := result[0].Interface().(fq.P)
	if !ok {
		return nil, fmt.Errorf("operator %s did not return a predicate", operator)
	}

	return predicate, nil
}

func parseArgs(fnType reflect.Type, value string) ([]reflect.Value, error) {
	parseValue := func(paramType reflect.Type, val string) (reflect.Value, error) {
		switch paramType.Kind() {
		case reflect.String:
			return reflect.ValueOf(val), nil
		case reflect.Float64:
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				return reflect.ValueOf(num), nil
			}
			return reflect.Value{}, fmt.Errorf("expected float64, got: %s", val)
		case reflect.Int:
			if num, err := strconv.Atoi(val); err == nil {
				return reflect.ValueOf(num), nil
			}
			return reflect.Value{}, fmt.Errorf("expected int, got: %s", val)
		case reflect.Interface:
			// For interface{}, try to parse as number first, then fall back to string
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				return reflect.ValueOf(num), nil
			}
			return reflect.ValueOf(val), nil
		default:
			return reflect.ValueOf(val), nil
		}
	}

	if fnType.IsVariadic() {
		parts := parseCommaSeparated(value)
		args := make([]reflect.Value, len(parts))
		for i, part := range parts {
			if num, err := strconv.ParseFloat(part, 64); err == nil {
				args[i] = reflect.ValueOf(num)
			} else {
				args[i] = reflect.ValueOf(part)
			}
		}
		return args, nil
	}

	parts := parseCommaSeparated(value)
	if len(parts) != fnType.NumIn() {
		return nil, fmt.Errorf("expected %d arguments, got %d", fnType.NumIn(), len(parts))
	}

	args := make([]reflect.Value, len(parts))
	for i, part := range parts {
		arg, err := parseValue(fnType.In(i), strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("argument %d: %v", i+1, err)
		}
		args[i] = arg
	}
	return args, nil
}

func parseCommaSeparated(value string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, char := range value {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
			fallthrough
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	for i, item := range result {
		if len(item) >= 2 && item[0] == '"' && item[len(item)-1] == '"' {
			result[i] = item[1 : len(item)-1]
		}
	}

	return result
}

func process(resultCh <-chan interface{}, srcErrCh, filterErrCh <-chan error, quiet bool) error {
	errorCh := make(chan error, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(errorCh)
		for {
			select {
			case err, ok := <-srcErrCh:
				if !ok {
					srcErrCh = nil
				} else {
					errorCh <- fmt.Errorf("source: %v", err)
					if !quiet {
						fmt.Fprintf(os.Stderr, "Source error: %v\n", err)
					}
				}
			case err, ok := <-filterErrCh:
				if !ok {
					filterErrCh = nil
				} else {
					errorCh <- fmt.Errorf("filter: %v", err)
					if !quiet {
						fmt.Fprintf(os.Stderr, "Filter error: %v\n", err)
					}
				}
			}
			if srcErrCh == nil && filterErrCh == nil {
				break
			}
		}
	}()

	outputErr := output(resultCh)

	<-done

	if outputErr != nil {
		return outputErr
	}
	select {
	case err := <-errorCh:
		return err
	default:
		return nil
	}
}

func output(resultCh <-chan interface{}) error {
	for result := range resultCh {
		bytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	}
	return nil
}

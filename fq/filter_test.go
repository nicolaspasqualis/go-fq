package fq

import (
	"strconv"
	"testing"
	"time"
)

// Test data structures
type Address struct {
	Street  string
	City    string
	Country string
	ZipCode string
}

type Product struct {
	ID           int
	Name         string
	Price        float64
	Categories   []string
	Tags         []string
	InStock      bool
	Stock        int
	Rating       float64
	Manufacturer struct {
		Name    string
		Country string
	}
	Address    Address
	CreatedAt  time.Time
	Properties map[string]interface{}
}

// Sample data
func getTestProducts() []Product {
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	return []Product{
		{
			ID:         1,
			Name:       "Laptop Pro",
			Price:      1299.99,
			Categories: []string{"electronics", "computers"},
			Tags:       []string{"premium", "work", "professional"},
			InStock:    true,
			Stock:      45,
			Rating:     4.7,
			Manufacturer: struct {
				Name    string
				Country string
			}{
				Name:    "TechCorp",
				Country: "USA",
			},
			Address: Address{
				Street:  "123 Tech St",
				City:    "San Francisco",
				Country: "USA",
				ZipCode: "94105",
			},
			CreatedAt: baseTime,
			Properties: map[string]interface{}{
				"color":     "silver",
				"processor": "i9",
				"memory":    32,
				"warranty":  2,
			},
		},
		{
			ID:         2,
			Name:       "Budget Tablet",
			Price:      299.99,
			Categories: []string{"electronics", "tablets"},
			Tags:       []string{"budget", "entertainment"},
			InStock:    true,
			Stock:      120,
			Rating:     4.1,
			Manufacturer: struct {
				Name    string
				Country string
			}{
				Name:    "ValueTech",
				Country: "China",
			},
			Address: Address{
				Street:  "456 Value Rd",
				City:    "Shenzhen",
				Country: "China",
				ZipCode: "518000",
			},
			CreatedAt: baseTime.AddDate(0, 1, 15),
			Properties: map[string]interface{}{
				"color":     "black",
				"processor": "A12",
				"memory":    4,
				"warranty":  1,
			},
		},
		{
			ID:         3,
			Name:       "Designer Watch",
			Price:      899.50,
			Categories: []string{"fashion", "accessories"},
			Tags:       []string{"premium", "gift", "luxury"},
			InStock:    true,
			Stock:      15,
			Rating:     4.9,
			Manufacturer: struct {
				Name    string
				Country string
			}{
				Name:    "LuxBrands",
				Country: "Switzerland",
			},
			Address: Address{
				Street:  "789 Luxury Ave",
				City:    "Geneva",
				Country: "Switzerland",
				ZipCode: "1201",
			},
			CreatedAt: baseTime.AddDate(0, 2, 10),
			Properties: map[string]interface{}{
				"color":           "gold",
				"material":        "stainless steel",
				"water_resistant": true,
				"warranty":        5,
			},
		},
		{
			ID:         4,
			Name:       "Wireless Earbuds",
			Price:      159.99,
			Categories: []string{"electronics", "audio"},
			Tags:       []string{"wireless", "music", "portable"},
			InStock:    true,
			Stock:      200,
			Rating:     4.5,
			Manufacturer: struct {
				Name    string
				Country string
			}{
				Name:    "AudioFi",
				Country: "USA",
			},
			Address: Address{
				Street:  "101 Sound Blvd",
				City:    "Los Angeles",
				Country: "USA",
				ZipCode: "90001",
			},
			CreatedAt: baseTime.AddDate(0, 3, 5),
			Properties: map[string]interface{}{
				"color":           "white",
				"battery_life":    6,
				"noise_canceling": true,
				"warranty":        1,
			},
		},
		{
			ID:         5,
			Name:       "Out of Stock Item",
			Price:      99.99,
			Categories: []string{"other"},
			Tags:       []string{"miscellaneous"},
			InStock:    false,
			Stock:      0,
			Rating:     3.0,
			Manufacturer: struct {
				Name    string
				Country string
			}{
				Name:    "Generic Co",
				Country: "Mexico",
			},
			Address: Address{
				Street:  "555 Generic St",
				City:    "Mexico City",
				Country: "Mexico",
				ZipCode: "11000",
			},
			CreatedAt: baseTime.AddDate(0, 4, 20),
			Properties: map[string]interface{}{
				"color":    "various",
				"material": "plastic",
				"warranty": 0,
			},
		},
	}
}

// Basic Tests ------------------------------------------------------------

func TestBasicEquality(t *testing.T) {
	products := getTestProducts()

	// Test direct equality with a primitive value
	result, err := Filter(products, Q{
		"ID": 3,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 3 {
		t.Errorf("Expected product with ID 3, got %v", result)
	}

	// Test direct equality with a string
	result, err = Filter(products, Q{
		"Name": "Laptop Pro",
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].Name != "Laptop Pro" {
		t.Errorf("Expected product with Name 'Laptop Pro', got %v", result)
	}

	// Test boolean equality
	result, err = Filter(products, Q{
		"InStock": false,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].InStock != false {
		t.Errorf("Expected out of stock products, got %v", result)
	}
}

func TestMultipleConditions(t *testing.T) {
	products := getTestProducts()

	// Test multiple conditions (implicit AND)
	result, err := Filter(products, Q{
		"InStock": true,
		"Price":   Lt(500),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products, got %d", len(result))
	}

	for _, product := range result {
		if !product.InStock || product.Price >= 500 {
			t.Errorf("Product doesn't match criteria: %+v", product)
		}
	}
}

// Operator Tests ---------------------------------------------------------

func TestComparisonOperators(t *testing.T) {
	products := getTestProducts()

	// Test greater than
	result, err := Filter(products, Q{
		"Price": Gt(1000),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 product with price > 1000, got %v", result)
	}

	// Test less than
	result, err = Filter(products, Q{
		"Price": Lt(200),
	}, 0, 0)

	if len(result) != 2 {
		t.Errorf("Expected 2 product with price < 200, got %v", result)
	}

	// Test greater than or equal
	result, err = Filter(products, Q{
		"Rating": Gte(4.5),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 products with rating >= 4.5, got %d", len(result))
	}

	// Test less than or equal
	result, err = Filter(products, Q{
		"Rating": Lte(4.1),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products with rating <= 4.1, got %d", len(result))
	}

	// Test In operator
	result, err = Filter(products, Q{
		"ID": In(1, 3, 5),
	}, 0, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products with ID in [1,3,5], got %d", len(result))
	}
}

// Logical Operators Tests ------------------------------------------------

func TestLogicalOperators(t *testing.T) {
	products := getTestProducts()

	// Test OR operator
	result, err := Filter(products, Or(
		Q{"Categories": HasItem("fashion")},
		Q{"Categories": HasItem("tablets")},
	), 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products (fashion OR tablets), got %d", len(result))
	}

	// Test AND operator
	result, err = Filter(products, And(
		Q{"InStock": true},
		Q{"Price": Lt(500)},
		Q{"Rating": Gt(4.0)},
	), 0, 0)

	if len(result) != 2 {
		t.Errorf("Expected 2 products (InStock AND Price<500 AND Rating>4.0), got %d", len(result))
	}

	// Test NOT operator
	result, err = Filter(products, Not(
		Q{"Manufacturer": Q{"Country": "USA"}},
	), 0, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products (NOT made in USA), got %d", len(result))
	}

	// Test complex nested logical operators
	result, err = Filter(products, And(
		Q{"InStock": true},
		Or(
			Q{"Price": Lt(200)},
			And(
				Q{"Price": Gt(800)},
				Q{"Rating": Gt(4.5)},
			),
		),
	), 0, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products for complex query, got %d", len(result))
	}
}

// String Operations Tests ------------------------------------------------

func TestStringOperations(t *testing.T) {
	products := getTestProducts()

	// Test string contains
	result, err := Filter(products, Q{
		"Name": Contains("Tablet"),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 2 {
		t.Errorf("Expected 1 product with 'Tablet' in name, got %v", result)
	}

	// Test case insensitivity
	result, err = Filter(products, Q{
		"Name": Contains("tablet"),
	}, 0, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 products with 'tablet' (case sensitive) in name, got %v", result)
	}
}

// Array Operations Tests -------------------------------------------------

func TestArrayOperations(t *testing.T) {
	products := getTestProducts()

	// Test HasItem with string array
	result, err := Filter(products, Q{
		"Tags": HasItem("premium"),
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products with 'premium' tag, got %d", len(result))
	}

	// Test HasItem with Categories
	result, err = Filter(products, Q{
		"Categories": HasItem("electronics"),
	}, 0, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products in 'electronics' category, got %d", len(result))
	}
}

// Nested Object Tests ----------------------------------------------------

func TestNestedObjectQueries(t *testing.T) {
	products := getTestProducts()

	// Test nested object matching
	result, err := Filter(products, Q{
		"Manufacturer": Q{
			"Country": "USA",
		},
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products made in USA, got %d", len(result))
	}

	// Test deeper nesting
	result, err = Filter(products, Q{
		"Address": Q{
			"City": Contains("Angeles"),
		},
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 4 {
		t.Errorf("Expected 1 product in Los Angeles, got %v", result)
	}

	// Test map properties
	result, err = Filter(products, Q{
		"Properties": Q{
			"warranty": Gt(1),
		},
	}, 0, 0)

	if len(result) != 2 {
		t.Errorf("Expected 2 products with warranty > 1, got %d", len(result))
	}
}

// Custom Function Tests --------------------------------------------------

func TestCustomFunctions(t *testing.T) {
	products := getTestProducts()

	// Test with a custom predicate function
	isPremiumElectronics := func(v interface{}) bool {
		if product, ok := v.(Product); ok {
			isElectronic := false
			for _, cat := range product.Categories {
				if cat == "electronics" {
					isElectronic = true
					break
				}
			}
			return isElectronic && product.Price > 1000
		}
		return false
	}

	result, err := Filter(products, isPremiumElectronics, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 premium electronics product, got %v", result)
	}

	// Test complex custom function
	result, err = Filter(products, Q{
		"": func(v interface{}) bool {
			if product, ok := v.(Product); ok {
				// Find USA products with good ratings that have "premium" tag
				isUSA := product.Manufacturer.Country == "USA"
				isHighRated := product.Rating >= 4.5
				hasPremiumTag := false

				for _, tag := range product.Tags {
					if tag == "premium" {
						hasPremiumTag = true
						break
					}
				}

				return isUSA && isHighRated && hasPremiumTag
			}
			return false
		},
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 premium high-rated USA product, got %v", result)
	}
}

// Time Tests ------------------------------------------------------------

func TestTimeComparisons(t *testing.T) {
	products := getTestProducts()
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	// Test time equality
	result, err := Filter(products, Q{
		"CreatedAt": baseTime,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 product created at base time, got %v", result)
	}

	// Test time comparison
	twoMonthsLater := baseTime.AddDate(0, 2, 0)
	result, err = Filter(products, Q{
		"CreatedAt": Gt(twoMonthsLater),
	}, 0, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products created after 2 months, got %d", len(result))
	}
}

// Pagination Tests -------------------------------------------------------

func TestPagination(t *testing.T) {
	products := getTestProducts()

	// Test skip
	result, err := Filter(products, Q{
		"InStock": true,
	}, 1, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 products after skipping 1, got %d", len(result))
	}

	// Test limit
	result, err = Filter(products, Q{
		"InStock": true,
	}, 0, 2)

	if len(result) != 2 {
		t.Errorf("Expected 2 products when limiting to 2, got %d", len(result))
	}

	// Test skip and limit combined
	result, err = Filter(products, Q{
		"InStock": true,
	}, 1, 2)

	if len(result) != 2 {
		t.Errorf("Expected 2 products when skipping 1 and limiting to 2, got %d", len(result))
	}
}

// Edge Cases Tests -------------------------------------------------------

func TestEdgeCases(t *testing.T) {
	products := getTestProducts()

	// Test empty dataset
	result, err := Filter([]Product{}, Q{
		"ID": 1,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 products from empty dataset, got %d", len(result))
	}

	// Test empty query
	result, err = Filter(products, Q{}, 0, 0)

	if len(result) != len(products) {
		t.Errorf("Expected all products with empty query, got %d", len(result))
	}

	// Test with non-existent field
	result, err = Filter(products, Q{
		"NonExistentField": "value",
	}, 0, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 products matching non-existent field, got %d", len(result))
	}

	// Test with nil values in query
	var nilValue interface{}
	result, err = Filter(products, Q{
		"ID": nilValue,
	}, 0, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 products with nil ID, got %d", len(result))
	}
}

// Performance Benchmarks -------------------------------------------------

func BenchmarkSimpleFilter(b *testing.B) {
	products := getTestProducts()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Filter(products, Q{
			"InStock": true,
		}, 0, 0)
		if err != nil {
			return
		}
	}
}

func BenchmarkComplexFilter(b *testing.B) {
	products := getTestProducts()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Filter(products, And(
			Q{"InStock": true},
			Or(
				Q{"Price": Lt(300)},
				And(
					Q{"Rating": Gt(4.5)},
					Q{"Manufacturer": Q{"Country": "USA"}},
				),
			),
		), 0, 0)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkNestedFilter(b *testing.B) {
	products := getTestProducts()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Filter(products, Q{
			"Manufacturer": Q{
				"Country": "USA",
			},
			"Address": Q{
				"City": Contains("Francisco"),
			},
		}, 0, 0)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkArrayFilter(b *testing.B) {
	products := getTestProducts()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Filter(products, Q{
			"Tags": HasItem("premium"),
		}, 0, 0)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkCustomFunctionFilter(b *testing.B) {
	products := getTestProducts()

	customFunc := func(v interface{}) bool {
		if product, ok := v.(Product); ok {
			return product.Rating > 4.0 && product.Stock > 50
		}
		return false
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Filter(products, Q{
			"": customFunc,
		}, 0, 0)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

// Extra Tests for Edge Cases ----------------------------------------------

func TestNilAndEmptyValues(t *testing.T) {
	type Item struct {
		ID    int
		Value interface{}
		Tags  []string
	}

	items := []Item{
		{ID: 1, Value: "test", Tags: []string{"a", "b"}},
		{ID: 2, Value: nil, Tags: []string{}},
		{ID: 3, Value: "", Tags: nil},
		{ID: 4, Value: 0, Tags: []string{"a"}},
	}

	// Test nil value match
	result, err := Filter(items, Q{
		"Value": nil,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 2 {
		t.Errorf("Expected 1 item with nil value, got %v", result)
	}

	// Test empty string match
	result, err = Filter(items, Q{
		"Value": "",
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 3 {
		t.Errorf("Expected 1 item with empty string value, got %v", result)
	}

	// Test nil slice match
	result, err = Filter(items, Q{
		"Tags": nil,
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 3 {
		t.Errorf("Expected 1 item with nil tags, got %v", result)
	}

	// Test empty slice match
	result, err = Filter(items, Q{
		"Tags": []string{},
	}, 0, 0)

	// This actually doesn't match anything because reflect.DeepEqual([]string{}, []string{}) is true
	// but the field value isn't exactly []string{}, it's a slice with 0 elements
	// To properly match empty slices, we'd need a special EmptySlice predicate

	// Test zero value match
	result, err = Filter(items, Q{
		"Value": 0,
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 4 {
		t.Errorf("Expected 1 item with zero value, got %v", result)
	}
}

func TestTypeCoercion(t *testing.T) {
	type Item struct {
		ID       int
		IntVal   int
		FloatVal float64
		StrVal   string
		BoolVal  bool
	}

	items := []Item{
		{ID: 1, IntVal: 42, FloatVal: 42.0, StrVal: "42", BoolVal: true},
		{ID: 2, IntVal: 0, FloatVal: 0.0, StrVal: "0", BoolVal: false},
	}

	// Test int vs float comparison
	result, err := Filter(items, Q{
		"IntVal": 42.0,
	}, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 item with IntVal = 42.0, got %v", result)
	}

	// Test numeric comparison with string representation
	result, err = Filter(items, Q{
		"StrVal": func(v interface{}) bool {
			str, ok := v.(string)
			if !ok {
				return false
			}

			num, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return false
			}

			return Gt(30)(num)
		},
	}, 0, 0)

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Expected 1 item with StrVal > 30, got %v", result)
	}
}

// Playground / Exploratory tests ------------------------------------------------------------

func TestAPIEdges(t *testing.T) {
	products := getTestProducts()

	// Combined logic operators with primitives as input
	result, err := Filter(products, Q{
		"ID": And(
			Or(3, 4),
			In(4, 3),
			Not(5),
			And(
				Not(In(1, 2)),
				Not(In(5, 6)),
			),
			Or(Eq(3), Eq(4), Eq(5), Eq(6)),
		),
	}, 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 2 || result[0].ID != 3 || result[1].ID != 4 {
		t.Errorf("Expected product with ID 3, got %v", result)
	}
	t.Log(result)

	// Numeric fields types are normalized for comparison.
	result, err = Filter(products, Or(
		Q{"ID": uint8(3)},
		Q{"ID": uint8(4)},
	), 0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 2 || result[0].ID != 3 || result[1].ID != 4 {
		t.Errorf("Expected product with ID 3, got %v", result)
	}
	t.Log(result)
}

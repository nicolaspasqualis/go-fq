# fq

Generic data filtering engine for Go with composable operators/queries and streaming support (zero dependencies). 
Works with any data structures - structs, maps, slices, or custom types.

```go
available := fq.Filter(servers, fq.And(
    fq.Q{
        "Status": fq.In("healthy", "degraded"),          // Multi-value match
        "Region": fq.Match("us-"),                       // Pattern matching  
        "CPU": fq.Lt(80),                                // Numeric comparison
        "Memory": fq.And(fq.Gt(20), fq.Lte(90)),         // Range filtering
        "Tags": fq.ContainsAll("web", "api"),            // Array operations
        "Location": fq.GeoWithin(40.7, -74.0, 50),       // Geospatial query
    },
    func(v interface{}) bool {                           // Custom predicate
        s := v.(Server)
        return len(s.ActiveConnections) < 1000 && s.Uptime.Hours() > 24
    },
), 0, 0)
```

## Core Concepts

fq filters any Go data using three composable query types:

- **`fq.Q`** (maps) for field-based matching
- **`fq.P`** (predicates) for custom logic  
- **Primitive values** for direct equality

These compose to create complex filtering logic:

```go
// Nested struct fields
type User struct {
    Profile struct {
        Location struct { Country string }
    }
}
users := []User{{Profile: {Location: {Country: "Canada"}}}}
canadians := fq.Filter(users, fq.Q{
    "Profile": fq.Q{
        "Location": fq.Q{"Country": "Canada"},
    },
}, 0, 0)

// Multiple conditions with logical operators
products := []Product{{Status: "active", Price: 150}}
fq.Filter(products, fq.Q{
    "Status": "active",
    "Price": fq.Or(fq.Eq(0), fq.Gt(100)),
    "Category": fq.In("electronics", "books"),
}, 0, 0)

// Custom predicates for complex logic
fq.Filter(products, func(v interface{}) bool {
    if p, ok := v.(Product); ok {
        return len(p.Tags) > 2 && p.Price < 1000
    }
    return false
}, 0, 0)
```

## Streaming API

Channel-based processing for large datasets:

```go
// Create data channel (from database, file, network, etc.)
dataCh := make(chan Product, 100)
go func() {
    for _, product := range hugeProductList {
        dataCh <- product
    }
    close(dataCh)
}()

// Stream filtering with constant memory usage
resultCh, errCh := fq.FilterC(dataCh, fq.Q{"Price": fq.Lt(100)}, 0, 0)

for product := range resultCh {
    // Process results as they arrive
}
```

## Available Operators

| Category | Operator | Description | Example |
|----------|----------|-------------|---------|
| **Logic** | `fq.And(...queries)` | All conditions must be true | `fq.And(fq.Q{"a": 1}, fq.Q{"b": 2})` |
| | `fq.Or(...queries)` | Any condition must be true | `fq.Or(fq.Q{"a": 1}, fq.Q{"b": 2})` |
| | `fq.Not(query)` | Negates a condition | `fq.Not(fq.Q{"status": "deleted"})` |
| **Comparison** | `fq.Gt(n)` | Greater than | `fq.Q{"price": fq.Gt(100)}` |
| | `fq.Gte(n)` | Greater than or equal | `fq.Q{"price": fq.Gte(100)}` |
| | `fq.Lt(n)` | Less than | `fq.Q{"price": fq.Lt(100)}` |
| | `fq.Lte(n)` | Less than or equal | `fq.Q{"price": fq.Lte(100)}` |
| | `fq.In(...)` | Value is in list | `fq.Q{"status": fq.In("active", "pending")}` |
| **String** | `fq.Contains(s)` | String contains substring | `fq.Q{"name": fq.Contains("Pro")}` |
| | `fq.Match(pattern)` | Case-insensitive match | `fq.Q{"name": fq.Match("john")}` |
| **Array** | `fq.HasItem(item)` | Array includes item | `fq.Q{"tags": fq.HasItem("urgent")}` |
| | `fq.ContainsAll(...)` | Array includes all items | `fq.Q{"tags": fq.ContainsAll("urgent", "important")}` |
| | `fq.ContainsAny(...)` | Array includes any item | `fq.Q{"tags": fq.ContainsAny("urgent", "important")}` |
| **Geospatial** | `fq.GeoWithin(lat, lng, radius)` | Coordinates within radius (km) | `fq.Q{"location": fq.GeoWithin(40.7, -74.0, 10)}` |

## Performance

No query optimization is applied - order filters from most to least restrictive:

```go
// Efficient: fast field check first
users := []User{{Active: true, Bio: "long bio..."}}
efficient := fq.And(
    fq.Q{"Active": true},           // fast boolean field
    fq.Q{"Bio": expensiveRegex},    // expensive operation on fewer items
)

// Inefficient: expensive check first
inefficient := fq.And(
    fq.Q{"Bio": expensiveRegex},    // runs on all items
    fq.Q{"Active": true},           // simple check after expensive work  
)
```

## Error Handling

```go
products := []Product{{Rating: 4.8}}
result, err := fq.Filter(products, fq.Q{"Rating": fq.Gte(4.5)}, 0, 0)
if err != nil {
    log.Printf("Filter error: %v", err)
}

// Streaming with error channels
dataCh := make(chan Product)
resultCh, errCh := fq.FilterC(dataCh, fq.Q{"Rating": fq.Gte(4.5)}, 0, 0)

go func() {
    for err := range errCh {
        log.Printf("Filter error: %v", err)
    }
}()
```

Custom predicates should handle type mismatches gracefully:

```go
safePredicate := func(v interface{}) bool {
    product, ok := v.(Product)
    if !ok {
        return false
    }
    return product.Rating >= 4.5
}
```

# CLI Tool

A command-line wrapper that provides JSONL input/output for the fq filtering engine:

```bash
# build
go build -o bin/fq ./cmd/fq

# filter JSONL files  
bin/fq data.jsonl "price:lt:500" "category:eq:electronics"
bin/fq data.jsonl "location:geowithin:40.7,-74.0,10"
bin/fq data.jsonl "tags:hasitem:urgent"
```

## CLI Usage

```bash
fq [options] <data-file> [filters...]
```

**Options:**
- `-skip <number>` - Skip first N results  
- `-limit <number>` - Limit to N results
- `-quiet` - Suppress error messages
- `-help` - Show help

**Filter syntax:** `field:operator:value`

**Operators:** `eq`, `gt`, `lt`, `gte`, `lte`, `match`, `contains`, `hasitem`, `in`, `geowithin`

*Note: The CLI wrapper handles JSONL parsing and output. The core fq library works with any Go data structures.*

--------

# ⚠ Status

#### Missing implementations
- Context handling API for cancellation, etc.
- Observability callbacks, hooks.
- Error definitions.

--------

# Installation

```bash
# library
go get github.com/nicolaspasqualis/go-fq

# CLI
go install github.com/nicolaspasqualis/go-fq/cmd/fq@latest
```

# Development

```bash
go build -o bin/fq ./cmd/fq
go test ./...
go run ./cmd/fq <data-file> [filters...]
```
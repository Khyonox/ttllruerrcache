# ttllruerrcache

Cache values with a TTL (and separate err cache)

# Example

```go
func ExampleCache() {
	c := Cache {
		ItemTTL: time.Second * 60,
	}
	c.Set("hello", "world")
	value, exists := c.Get("hello")
	fmt.Println(value, exists)
	// Output: world true
}
```

The cache also works with default values


```go
func ExampleCache_Defaults() {
	var c Cache
	c.Set("hello", "world")
	value, exists := c.Get("hello")
	fmt.Println(value, exists)
	// Output: world true
}
```
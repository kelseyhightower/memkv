# memkv

Simple in memory k/v store.


## Usage

```
package main

import (
	"fmt"

	"github.com/kelseyhightower/memkv"
)

func main() {
	s := memkv.New()
	s.Set("/myapp/database/username", "admin")
	s.Set("/myapp/database/password", "123456789")
	s.Set("/myapp/port", "80")

	// Get a specific node.
	node, ok := s.Get("/myapp/database/username")	
	if ok {
		fmt.Printf("Key: %s, Value: %s\n", node.Key, node.Value)
	}

	// Get all the nodes that where Key matches pattern.
	nodes, err := s.GetAll("/myapp/*/*")
	if err == nil {
		for _, n := range nodes {
			fmt.Printf("Key: %s, Value: %s\n", n.Key, n.Value)
		}
	}
}	
```

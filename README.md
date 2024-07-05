# APIS
The MakeRequest function is a utility for performing customizable HTTP requests in Go. It allows setting headers, query parameters, and handling errors effectively. Additionally, it provides a Response structure to facilitate reading and decoding the server response.

```bash
    go get -u github.com/user0608/apis
```
```go
package main

import (
    "context"
    "fmt"
    "log"
)

type Data struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Value string `json:"value"`
}

func main() {
    ctx := context.Background()
    apiURL := "https://api.example.com/data"

    response := MakeRequest(
        ctx,
        apiURL,
        WithHeader("Authorization", "Bearer token"),
        WithHeader("Accept", "application/json"),
        WithQueryParam("id", "123"),
    )
    if response.Err != nil {
        log.Fatalf("Request failed: %v", response.Err)
    }

    var data Data
    if err := response.Scan(&data); err != nil {
        log.Fatalf("Failed to decode response: %v", err)
    }

    fmt.Printf("Status Code: %d\n", response.StatusCode)
    fmt.Printf("Response Data: %+v\n", data)
}
```
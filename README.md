# A http client for rest request

 `surport chainable usage`

## Examples

```go
package main

import (
	hc "github.com/zrcoder/httpclient"
	"net/http"
)

func main() {
	person := struct {
		Age  int
		Name string
	}{Age: 27, Name: "Tom"}

	hc.New().
		Post("http://127.0.0.1:8888/test").
		Header("some key", "some value").
		ContentType(hc.ContentTypeJson).
		Body(person).
		Do(func(response *http.Response, body []byte, err error) {
		// do something with response
	})
}
```

or

```go
client := hc.New().
    Post("http://127.0.0.1:8888/test").
    Header("some key", "some value").
    ContentType(hc.ContentTypeJson).
    Body(person)
resp, body, err := client.Go()
...
```

you can review at test cases to see more exapmples
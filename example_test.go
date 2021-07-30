package httpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
)

func Example() {
	var persons []Person
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()
		body, _ := ioutil.ReadAll(r.Body)

		switch r.Method {
		case GET:
			if data, err := json.Marshal(persons); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write(nil)
			} else {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(data)
			}
		case PUT:
			var p Person
			err := json.Unmarshal(body, &p)
			if err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				_, _ = w.Write([]byte("body in request is not a person"))
				return
			}
			persons = append(persons, p)
			w.WriteHeader(http.StatusOK)
		case POST:
			var p Person
			err := json.Unmarshal(body, &p)
			if err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				_, _ = w.Write([]byte("body in request is not a person"))
				return
			}
			indexToReplace := -1
			for index, pp := range persons {
				if pp.Name == p.Name {
					indexToReplace = index
					break
				}
			}
			if indexToReplace > -1 {
				persons[indexToReplace] = p
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("the person not found"))
		case DELETE:
			var p Person
			err := json.Unmarshal(body, &p)
			if err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				_, _ = w.Write([]byte("body in request is not a person"))
				return
			}
			var newPersons []Person
			for _, pp := range persons {
				if pp.Name != p.Name {
					newPersons = append(newPersons, pp)
				}
			}
			if len(persons)-1 == len(newPersons) {
				persons = newPersons
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("the person not found"))
				return
			}
		}
	}))
	defer server.Close()

	dog := Pet{Name: "wangwang", Color: "black"}
	tom := Person{Name: "Tom", Age: 27, Pet: dog}

	cat := Pet{Name: "miumu", Color: "white"}
	joe := Person{Name: "Joe", Age: 3, Pet: cat}

	addPerson(server.URL+"/add", tom)
	addPerson(server.URL+"/add", joe)
	queryPersons(server.URL + "/persons")
	modifyPersonAge(server.URL+"/modify", joe, 5)
	removePerson(server.URL+"/remove", tom)
	queryPersons(server.URL + "/persons")

	// Output:
	// quiried persons: [{27 Tom {wangwang black}} {3 Joe {miumu white}}]
	// quiried persons: [{5 Joe {miumu white}}]
}

func addPerson(url string, p Person) {
	New().Put(url).Body(p).Do(func(response *http.Response, err error) {
		if err != nil {
			fmt.Println("add a person failed:", err)
			return
		}
		defer func() {
			_ = response.Body.Close()
		}()
	})
}
func queryPersons(url string) {
	New().Get(url).Do(func(response *http.Response, err error) {
		if err != nil {
			fmt.Println("query persons failed, status code:", err)
			return
		}
		defer func() {
			_ = response.Body.Close()
		}()
		if response.StatusCode != http.StatusOK {
			fmt.Println("query persons status is:", response.Status)
		}
		var quariedPersons []Person
		body, _ := ioutil.ReadAll(response.Body)
		err = json.Unmarshal(body, &quariedPersons)
		if err != nil {
			fmt.Println("unmarshal resonse as a person array failed:", err)
			return
		}
		fmt.Println("quiried persons:", quariedPersons)
	})
}

func modifyPersonAge(url string, p Person, age int) {
	p.Age = age
	_, _ = New().Post(url).Body(p).Go()
}

func removePerson(url string, p Person) {
	_, _ = New().Delete(url).Body(p).Go()
}

var (
	// These certs are just the server of httptest.NewTLSServer() used
	// see the internal package "net/http/internal"

	testCAContent = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)
	testCertContent = testCAContent
	testKeyContent  = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)
)

func ExampleTlsConfig() {
	const url = "/test"
	const body = "Hello world!"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == url {
			_, _ = w.Write([]byte(body))
		}
	}))
	defer server.Close()

	New().
		AddCAContent(testCAContent).
		AddCertContent(testCertContent, testKeyContent).
		Get(server.URL + url).
		InsecureSkipVerify(true).
		Do(func(response *http.Response, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				body, _ := ioutil.ReadAll(response.Body)
				fmt.Println(string(body))
				_ = response.Body.Close()
			}
		})

	// output:
	// Hello world!
}

func ExampleQuery() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.RawQuery))
	}))
	defer server.Close()

	c := New().Get(server.URL)
	c.AppendQueries(map[string]string{
		"a": "hello",
		"b": "world",
	})
	c.AppendQuery("c", "hi")
	c.AppendQuery("d", "you")
	resp, err := c.Go()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	queries := strings.Split(string(body), "&")
	sort.Strings(queries)
	for _, s := range queries {
		fmt.Println(s)
	}

	// Output:
	// a=hello
	// b=world
	// c=hi
	// d=you
}

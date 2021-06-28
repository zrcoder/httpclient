package httpclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Pet struct {
	Name  string
	Color string
}

type Person struct {
	Age  int
	Name string
	Pet  Pet
}

func TestGet(t *testing.T) {
	cases := []struct {
		url         string
		headerKey   string
		headerValue string
		respBody    string
		respStatus  int
	}{
		{url: "/", respBody: "Hello world", respStatus: http.StatusAccepted},
		{url: "/example", headerKey: "Key", headerValue: "Value", respBody: "Example body", respStatus: http.StatusAccepted},
		{url: "/example/ttt", headerKey: "kkk", headerValue: "vvv", respBody: "Example ttt body", respStatus: http.StatusOK},
	}

	for _, c := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != GET {
				t.Errorf("Expected method %s; got %s", GET, r.Method)
			}
			w.WriteHeader(c.respStatus)
			_, _ = w.Write([]byte(c.respBody))

			if c.headerKey != "" {
				v := r.Header.Get(c.headerKey)
				if v != c.headerValue {
					t.Errorf("get error header value, key is %s, expected value is %s, but got %s", c.headerKey, c.headerValue, v)
				}
			}
		}))
		defer server.Close()

		callback := func(response *http.Response, body []byte, err error) {
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != c.respStatus {
				t.Error("response status is not", c.respStatus)
			}
			if string(body) != c.respBody {
				t.Errorf("response body should be %s, got %s", c.respBody, string(body))
			}
		}

		if c.headerKey == "" {
			New().Get(server.URL + c.url).Do(callback)
		} else {
			New().Get(server.URL+c.url).
				Header(c.headerKey, c.headerValue).
				Do(callback)
		}
	}
}

func TestPost(t *testing.T) {
	const rootUrl = "/"
	const headerUrl = "header"
	const headerKey = "post key"
	const headerValue = "post value"
	const jsonUrl = "/body/json/changepet"
	const jsonRespPetName = "wangwang"
	const jsonReqPersonName = "Joe"
	const mapUrl = "/body/map"
	const mapKey = "ttKey"
	const mapValue = "ttValue"
	const mapStr = `{"ttKey":"ttValue"}`
	const arrayUrl = "/body/array"
	var arrayBody = []int{5, 8}
	const textUrl = "/body/text"
	const textData = "Hello world!"
	const intBody = 3
	const intUrl = "/body/int"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case rootUrl:
			t.Log("root url")
		case headerUrl:
			if r.Header.Get(headerKey) != headerValue {
				t.Errorf("expected header value of key %s == %s; got %s", headerKey, headerValue, r.Header.Get(headerKey))
			}
		case jsonUrl:
			defer func() {
				_ = r.Body.Close()
			}()
			body, _ := ioutil.ReadAll(r.Body)
			p := Person{}
			_ = json.Unmarshal(body, &p)
			if p.Name == jsonReqPersonName {
				p.Pet = Pet{
					Name: jsonRespPetName,
				}
			}
			resp, _ := json.Marshal(&p)
			_, _ = w.Write(resp)
		case mapUrl, arrayUrl, textUrl, intUrl:
			defer func() {
				_ = r.Body.Close()
			}()
			body, _ := ioutil.ReadAll(r.Body)
			_, _ = w.Write(body)
		}
	}))
	defer server.Close()

	_, _, _ = New().Post(server.URL + "/").GoStr()
	_, _, _ = New().Post(server.URL+headerUrl).Header(headerKey, headerValue).GoStr()

	cat := Pet{
		Name:  "Miumiu",
		Color: "white",
	}
	joe := Person{
		Name: "Joe",
		Age:  27,
		Pet:  cat,
	}

	New().Post(server.URL + jsonUrl).
		ContentType(ContentTypeJson).
		Body(joe).
		Do(func(resp *http.Response, body []byte, err error) {
			if err != nil {
				t.Fatal(err)
			}
			p := Person{}
			err = json.Unmarshal(body, &p)
			if err != nil {
				t.Error("response is not a person struct!")
			}
			if p.Pet.Name != jsonRespPetName {
				t.Error("response pet name not expected")
			}
		})

	New().Post(server.URL + mapUrl).
		Body(map[string]string{mapKey: mapValue}).
		DoStr(func(response *http.Response, body string, err error) {
			if err != nil {
				t.Fatal(err)
			}
			if mapStr != body {
				t.Error("inputed map is not the one of response string")
			}
		})

	New().Post(server.URL + arrayUrl).
		Body(arrayBody).
		Do(func(response *http.Response, body []byte, err error) {
			if err != nil {
				t.Fatal(err)
			}
			var array []int
			err = json.Unmarshal(body, &array)
			if err != nil {
				t.Error(err)
			}
			if len(array) != len(arrayBody) {
				t.Errorf("expected array length %d, but got %d\n", len(arrayBody), len(array))
			} else {
				for i := 0; i < len(array); i++ {
					if array[i] != arrayBody[i] {
						t.Errorf("expected element in array is %d, got %d", arrayBody[i], array[i])
					}
				}
			}
		})

	New().Post(server.URL + textUrl).Body(textData).DoStr(func(response *http.Response, body string, err error) {
		if err != nil {
			t.Fatal(err)
		}
		if body != textData {
			t.Errorf("expect text respons: %s, got: %s", textData, body)
		}
	})

	New().Post(server.URL + intUrl).Body(intBody).DoStr(func(response *http.Response, body string, err error) {
		if err != nil {
			t.Fatal(err)
		}
		t.Log("body is a number, response is:", body)
	})
}

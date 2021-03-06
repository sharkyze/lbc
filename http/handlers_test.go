package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sharkyze/lbc/metrics"
)

type httpTestRequest struct {
	method string
	params map[string]string
}

func mustDoTestRequest(t *testing.T, u string, req httpTestRequest) *http.Response {
	r, err := http.NewRequest(req.method, u, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}

	q := url.Values{}

	for key, value := range req.params {
		q.Add(key, value)
	}

	r.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return res
}

func Test_handlers_handleFizzBuzz(t *testing.T) {
	type want struct {
		status        int
		CheckResponse bool
		response      response
	}

	tests := map[string]struct {
		args httpTestRequest
		want want
	}{
		"method not allowed": {
			args: httpTestRequest{
				method: http.MethodPost,
				params: nil,
			},
			want: want{
				status: http.StatusMethodNotAllowed,
			},
		},
		"missing int1": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "",
					"int2":  "5",
					"limit": "100",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"error parsing int1": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "error",
					"int2":  "5",
					"limit": "100",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"missing int2": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "",
					"limit": "100",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"error parsing int2": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "error",
					"limit": "100",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"missing limit": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "5",
					"limit": "",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"error parsing limit": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "5",
					"limit": "error",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"missing str1": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "5",
					"limit": "100",
					"str1":  "",
					"str2":  "Buzz",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"missing str2": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "5",
					"limit": "100",
					"str1":  "Fizz",
					"str2":  "",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
		"ok": {
			args: httpTestRequest{
				method: http.MethodGet,
				params: map[string]string{
					"int1":  "3",
					"int2":  "5",
					"limit": "10",
					"str1":  "Fizz",
					"str2":  "Buzz",
				},
			},
			want: want{
				status:        http.StatusOK,
				CheckResponse: true,
				response: response{
					Data: []interface{}{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz"},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := log.New(os.Stdout, "tests", log.Ltime)
			m := metrics.NewInMemoryMetrics()
			hs := newHandlers(l, &m)

			ts := httptest.NewServer(http.HandlerFunc(hs.handleFizzBuzz))
			defer ts.Close()

			res := mustDoTestRequest(t, ts.URL, tt.args)
			defer res.Body.Close()

			if tt.want.status != res.StatusCode {
				t.Errorf("status code error")
			}

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("error reading response body: %v", err)
			}

			if tt.want.CheckResponse {
				var r response

				err = json.Unmarshal(b, &r)
				if err != nil {
					t.Errorf("error unmarshalling response json: %v", err)
				}

				if !cmp.Equal(r, tt.want.response) {
					t.Errorf(cmp.Diff(r, tt.want.response))
				}

				fmt.Println(r)
			}
		})
	}
}

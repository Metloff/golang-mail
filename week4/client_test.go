package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Users struct {
	XMLName xml.Name `xml:"root"`
	Users   []Row    `xml:"row"`
}

type Row struct {
	XMLName   xml.Name `xml:"row"`
	ID        int      `xml:"id"`
	GuID      string   `xml:"guid"`
	IsActive  bool     `xml:"isActive"`
	Balance   string   `xml:"balance"`
	Picture   string   `xml:"picture"`
	Age       int      `xml:"age"`
	EyeColor  string   `xml:"eyeColor"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	Gender    string   `xml:"gender"`
	Company   string   `xml:"company"`
	Email     string   `xml:"email"`
	Phone     string   `xml:"phone"`
	Address   string   `xml:"address"`
	About     string   `xml:"about"`
}

type TestCase struct {
	SearchRequest SearchRequest
	Result        *SearchResponse
	IsError       bool
	ResultErr     error
}

var users Users

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if len(users.Users) < 1 {
		prepareDB()
	}

	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		panic(err)
	}

	query := r.FormValue("query")

	switch query {
	case "good_case":
		resUsers, _ := json.Marshal(users.Users[:limit])

		w.WriteHeader(http.StatusOK)
		w.Write(resUsers)
	case "forever_one_result":
		resUsers, _ := json.Marshal(users.Users[:1])

		w.WriteHeader(http.StatusOK)
		w.Write(resUsers)
	case "timeout":
		time.Sleep(2 * time.Second)

		io.WriteString(w, `{}`)
	case "broken_json":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 400`) //broken json
	case "unauthorized":
		w.WriteHeader(http.StatusUnauthorized)
	case "internal_error":
		w.WriteHeader(http.StatusInternalServerError)
	case "bad_request1":
		resp, _ := json.Marshal(SearchErrorResponse{})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
	default:
	}
}

func prepareDB() {
	// Open our xmlFile
	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Println(err)
	}
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	xml.Unmarshal(byteValue, &users)
}

func TestTimeout(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "timeout",
			},
			IsError:   true,
			ResultErr: fmt.Errorf("timeout for"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if err != nil && !strings.HasPrefix(err.Error(), testCase.ResultErr.Error()) {
			t.Errorf("Expect: %#v. Got: %#v", testCase.ResultErr, err)
		}
		if err != nil && result != nil {
			t.Errorf("Got error and not nil result. Error: %#v. Result: %#v.", err, result)
		}
	}

	ts.Close()
}

func TestStatusCode(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "unauthorized",
			},
			IsError:   true,
			ResultErr: fmt.Errorf("Bad AccessToken"),
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "internal_error",
			},
			IsError:   true,
			ResultErr: fmt.Errorf("SearchServer fatal error"),
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "bad_request1",
			},
			IsError:   true,
			ResultErr: fmt.Errorf("unknown bad request error"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if err != nil && !strings.HasPrefix(err.Error(), testCase.ResultErr.Error()) {
			t.Errorf("Expect: %#v. Got: %#v", testCase.ResultErr, err)
		}
		if err != nil && result != nil {
			t.Errorf("Got error and not nil result. Error: %#v. Result: %#v.", err, result)
		}
	}

	ts.Close()
}

func TestIncorrectjsonInJsonResponse(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "broken_json",
			},
			IsError:   true,
			ResultErr: fmt.Errorf("cant unpack result json:"),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if err != nil && !strings.HasPrefix(err.Error(), testCase.ResultErr.Error()) {
			t.Errorf("Expect: %#v. Got: %#v", testCase.ResultErr, err)
		}
		if err != nil && result != nil {
			t.Errorf("Got error and not nil result. Error: %#v. Result: %#v.", err, result)
		}
	}

	ts.Close()
}

func TestResponseFromRemoteServerIsLessThanLimit(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 20,
				Query: "forever_one_result",
			},
			IsError: false,
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 10,
				Query: "forever_one_result",
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if result != nil && len(result.Users) != 1 {
			t.Errorf("Got incorrect number of users. Expect: 1. Got: %v ", len(result.Users))
		}
	}

	ts.Close()
}

func TestLimitGreaterThan25(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 25,
				Query: "good_case",
			},
			IsError: false,
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 26,
				Query: "good_case",
			},
			IsError: false,
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 40,
				Query: "good_case",
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if result != nil && len(result.Users) != 25 {
			t.Errorf("Got incorrect number of users. Expect: 25. Got: %v ", len(result.Users))
		}
	}

	ts.Close()
}

func TestLimitIsLessThanZero(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Limit: -1,
				Query: "good_case",
			},
			Result:    nil,
			IsError:   true,
			ResultErr: fmt.Errorf("limit must be > 0"),
		},
		TestCase{
			SearchRequest: SearchRequest{
				Limit: 0,
				Query: "good_case",
			},
			Result:  nil,
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if err != nil && testCase.ResultErr.Error() != err.Error() {
			t.Errorf("Expect: %#v. Got: %#v", testCase.ResultErr, err)
		}
		if err != nil && result != nil {
			t.Errorf("Got error and not nil result. Error: %#v. Result: %#v.", err, result)
		}
	}

	ts.Close()
}

func TestOffsetIsLessThanZero(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: SearchRequest{
				Offset: -1,
				Query:  "good_case",
			},
			Result:    nil,
			IsError:   true,
			ResultErr: fmt.Errorf("offset must be > 0"),
		},
		TestCase{
			SearchRequest: SearchRequest{
				Offset: 0,
				Query:  "good_case",
			},
			Result:  nil,
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{URL: ts.URL}

	for _, testCase := range cases {
		result, err := c.FindUsers(testCase.SearchRequest)
		if err != nil && !testCase.IsError {
			t.Errorf("Unexpected error: %#v", err)
		}
		if err == nil && testCase.IsError {
			t.Errorf("Expected error, got nil")
		}
		if err != nil && testCase.ResultErr.Error() != err.Error() {
			t.Errorf("Expect: %#v. Got: %#v", testCase.ResultErr, err)
		}
		if err != nil && result != nil {
			t.Errorf("Got error and not nil result. Error: %#v. Result: %#v.", err, result)
		}
	}

	ts.Close()
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func main() {
	FastSearch(ioutil.Discard)
}

type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var user User
	seenBrowsers := make([]string, 0, 2)
	foundUsers := ""

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		isAndroid := false
		isMSIE := false

		err = json.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			panic(err)
		}

		for _, browser := range user.Browsers {
			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
				}
			}

			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		user.Email = strings.ReplaceAll(user.Email, "@", " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, user.Email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func StdJson() {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(fileContents), "\n")

	users := make([]*User, 0)
	for _, line := range lines {
		user := &User{}
		err := json.Unmarshal([]byte(line), user)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}
}

func FastJson() {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(fileContents), "\n")
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	users := make([]*User, 0)
	for _, line := range lines {
		user := &User{}
		err := json.Unmarshal([]byte(line), user)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}
}

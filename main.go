package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type paramCheck struct {
	url   string
	param string
}

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: time.Second,
	}).DialContext,
}

var httpClient = &http.Client{
	Transport: transport,
	Timeout:   15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	initialChecks := make(chan paramCheck, 40)

	// Step 1: Filter URLs to see which parameters are actually reflected in the response
	appendChecks := makePool(initialChecks, func(c paramCheck, output chan paramCheck) {
		reflected, err := getReflectedParams(c.url)
		if err != nil {
			return
		}

		if len(reflected) == 0 {
			return
		}

		for _, param := range reflected {
			output <- paramCheck{c.url, param}
		}
	})

	// Step 2: For each reflected parameter, test the special character set
	done := makePool(appendChecks, func(c paramCheck, output chan paramCheck) {
		specialChars := []string{"\"", "'", "<", ">", "$", "|", "(", ")", "`", ":", ";", "{", "}"}
		var unfiltered []string

		for _, char := range specialChars {
			// We use a unique prefix/suffix to ensure we aren't seeing a false positive
			wasReflected, err := checkAppend(c.url, c.param, "kxss"+char+"test")
			if err == nil && wasReflected {
				unfiltered = append(unfiltered, char)
			}
		}

		if len(unfiltered) > 0 {
			fmt.Printf("URL: %s Param: %s Unfiltered: %v\n", c.url, c.param, unfiltered)
		}
	})

	// Read from Stdin
	for sc.Scan() {
		initialChecks <- paramCheck{url: sc.Text()}
	}

	close(initialChecks)
	<-done
}

// getReflectedParams checks the initial URL to see which query keys appear in the body
func getReflectedParams(targetURL string) ([]string, error) {
	out := make([]string, 0)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return out, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) kxss/1.1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	// Skip redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return out, nil
	}

	// Only process HTML responses
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(strings.ToLower(ct), "html") {
		return out, nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	body := string(b)

	u, err := url.Parse(targetURL)
	if err != nil {
		return out, err
	}

	for key, vv := range u.Query() {
		for _, v := range vv {
			if v != "" && strings.Contains(body, v) {
				out = append(out, key)
				break 
			}
		}
	}

	return out, nil
}

// checkAppend appends a specific character/string to a parameter and checks for reflection
func checkAppend(targetURL, param, suffix string) (bool, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	qs := u.Query()
	val := qs.Get(param)
	qs.Set(param, val+suffix)
	u.RawQuery = qs.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) kxss/1.1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(b), suffix), nil
}

func makePool(input chan paramCheck, fn func(paramCheck, chan paramCheck)) chan paramCheck {
	var wg sync.WaitGroup
	output := make(chan paramCheck)

	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			for c := range input {
				fn(c, output)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

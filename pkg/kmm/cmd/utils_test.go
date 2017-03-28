package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestGetHostNamesFromUrls(t *testing.T) {
	hosts, err := GetHostNamesFromUrls("http://localhost:2380", []string{})
	if err != nil {
		t.Fatal(err)
	}
	if hosts != nil {
		if hosts[0] != "localhost" {
			t.Errorf("Expecting \"localhost\" but got %v", hosts[0])
		}
	} else {
		t.Errorf("Expecting []string{\"localhost\"} but got %v", hosts)
	}
}

func TestGetUrlsFromInitialClusterString(t *testing.T) {

	urls, err := GetUrlsFromInitialClusterString("")
	if len(urls) -1 > 0 {
		t.Errorf("Didn't expect any URL's returned for empty string input but got:%q.", urls)
	}
	err = testAString(
		"10.111.2.117=https://10.111.2.117:2380",
		"https://10.111.2.117:2380",
		1)
	if err != nil {
		t.Fatal(err)
	}

	testAString(
		"10.111.2.117=https://10.111.2.117:2380,10.111.2.149=https://10.111.2.149:2380,10.111.2.188=https://10.111.2.188:2380",
		"https://10.111.2.117:2380,https://10.111.2.149:2380,https://10.111.2.188:2380",
		3)
	if err != nil {
		t.Fatal(err)
	}

	err = testAString(
		"localhost=http://localhost:2380",
		"http://localhost:2380",
		1)
	if err != nil {
		t.Fatal(err)
	}

	err = testAString(
		"node1=http://10.0.0.1:2380",
		"http://10.0.0.1:2380",
		1)
	if err != nil {
		t.Fatal(err)
	}

}

func testAString(initialString string, expectedString string, expectdNumber int) (error) {
	urls, err := GetUrlsFromInitialClusterString(initialString)
	if err != nil {
		return err
	}
	if urls != expectedString {
		return fmt.Errorf("Expected %q but got %q", expectedString, urls)
	}
	urlsAry := strings.Split(urls, ",")
	numUrls := len(urlsAry)
	if numUrls != expectdNumber {
		return fmt.Errorf("Expected %d URL's returned for string %q.", numUrls, urls)
	}
	for _, s := range urlsAry {
		_, err := url.Parse(s)
		if err != nil {
			return err
		}
	}
	return nil
}

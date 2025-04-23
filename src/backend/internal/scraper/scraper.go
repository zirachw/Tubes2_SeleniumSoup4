package scraper

import (
	"fmt"
	"io"
	"net/http"
)

// coba doang (TOLONG HAPUS)
func Hello() {
	fmt.Println("Hello, world!")
}

// coba doang (TOLONG HAPUS)
func TestRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

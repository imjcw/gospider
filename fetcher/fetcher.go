package fetcher

import (
	"bufio"
	"errors"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Fetch .
func Fetch(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New("http status code failed " + strconv.Itoa(resp.StatusCode))
	}
	reader := bufio.NewReader(resp.Body)
	encoding := determineEncoding(reader)
	utf8Reader := transform.NewReader(reader, encoding.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// determineEncoding determine encoding
func determineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}

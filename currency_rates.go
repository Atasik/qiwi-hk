package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	YYYYMMDD = "2006-01-02"
	DDMMYYYY = "02-01-2006"
	timeout  = 5 * time.Second
	path     = "http://www.cbr.ru/scripts/XML_daily.asp?date_req=%s"
)

type Curs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Valute  []struct {
		CharCode string `xml:"CharCode"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	}
}

type Client struct {
	client http.Client
}

func newClient(timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		return nil, errors.New("invalid timeout")
	}

	return &Client{
		client: http.Client{
			Timeout: timeout,
		},
	}, nil
}

func main() {
	var code string
	var dateStr string

	flag.StringVar(&code, "code", "USD", "currency code")
	flag.StringVar(&dateStr, "date", "2022-10-08", "date")

	flag.Parse()

	dateStr, err := formatDate(dateStr)
	if err != nil {
		fmt.Print(err)
		return
	}

	c, err := newClient(timeout)
	if err != nil {
		fmt.Print(err)
		return
	}

	result, err := c.getCurs(dateStr)
	if err != nil {
		print("failed to get curs")
		return
	}

	found := false
	for _, v := range result.Valute {
		if v.CharCode == code {
			fmt.Printf("%s (%s) %s", code, v.Name, v.Value)
			found = true
		}
	}
	if !found {
		fmt.Print("nothing found")
	}
}

func formatDate(dateStr string) (string, error) {
	date, err := time.Parse(YYYYMMDD, dateStr)
	if err != nil {
		return "", errors.New("date invalid format")
	}

	newDate := date.Format(DDMMYYYY)

	dateStr = strings.Replace(newDate, "-", "/", 2)
	return dateStr, nil
}

func (c *Client) getCurs(dateStr string) (Curs, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(path, dateStr), nil)
	if err != nil {
		return Curs{}, err
	}
	req.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	var result Curs
	if xmlBytes, err := c.getXML(req); err != nil {
		return Curs{}, err
	} else {
		if err := decodeXML(xmlBytes, &result); err != nil {
			return Curs{}, err
		}
	}

	return result, err
}

func decodeXML(xmlBytes []byte, v any) error {
	decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
	decoder.CharsetReader = charset
	if err := decoder.Decode(&v); err != nil {
		fmt.Printf("[ERROR] Cannot decode file %e", err)
		return err
	}
	return nil
}

func (c *Client) getXML(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func charset(charset string, input io.Reader) (io.Reader, error) {
	switch charset {
	case "windows-1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	default:
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}
}

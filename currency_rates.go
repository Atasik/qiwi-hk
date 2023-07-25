package main

import (
	"bytes"
	"encoding/xml"
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

type userSession struct {
	client http.Client
}

func main() {
	var code string
	var dateStr string

	flag.StringVar(&code, "code", "USD", "currency code")
	flag.StringVar(&dateStr, "date", "2022-10-08", "date")

	flag.Parse()

	// code = "USD"
	// dateStr = "2002-03-02"
	date, err := time.Parse(YYYYMMDD, dateStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	newDate := date.Format(DDMMYYYY)

	dateStr = strings.Replace(newDate, "-", "/", 2)

	u := &userSession{
		client: http.Client{
			Timeout: timeout,
		},
	}

	var result ValCurs

	req, err := http.NewRequest("GET", fmt.Sprintf(path, dateStr), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)
	if xmlBytes, err := u.getXML(req); err != nil {
		fmt.Printf("Failed to get XML: %v", err)
		return
	} else {
		decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
		decoder.CharsetReader = charset
		if err := decoder.Decode(&result); err != nil {
			fmt.Printf("[ERROR] Cannot decode file %e", err)
			return
		}
	}

	for _, v := range result.Valute {
		if v.CharCode == code {
			fmt.Printf("%s (%s) %s", code, v.Name, v.Value)
		}
	}
}

func (u *userSession) getXML(req *http.Request) ([]byte, error) {
	resp, err := u.client.Do(req)
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

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Valute  []struct {
		CharCode string `xml:"CharCode"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	}
}

func charset(charset string, input io.Reader) (io.Reader, error) {
	switch charset {
	case "windows-1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	default:
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}
}

package baidu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type BCMClient struct {
	ak     string
	sk     string
	host   string
	method string
	http   *http.Client
}

func NewBCMClient(AK, SK, host string) *BCMClient {
	return &BCMClient{
		ak:     AK,
		sk:     SK,
		host:   host,
		method: http.MethodGet,
		http:   &http.Client{},
	}
}

type QueryParam struct {
	K, V string
}

func (c *BCMClient) Send(path string, queryList []QueryParam) (map[string]interface{}, error) {
	timeStamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	authStringPrefix := fmt.Sprintf("bce-auth-v1/%s/%s/10000", c.ak, timeStamp)
	h := hmac.New(sha256.New, []byte(c.sk))
	h.Write([]byte(authStringPrefix))
	signKey := hex.EncodeToString(h.Sum(nil))

	header := map[string]string{
		"host":       c.host,
		"x-bce-date": timeStamp,
	}

	var keys []string
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(header))
	for _, key := range keys {
		if value, ok := header[key]; ok {
			tempStr := url.QueryEscape(strings.ToLower(key)) + ":" + url.QueryEscape(value)
			result = append(result, tempStr)
		}
	}

	query := c.buildQueryParams(queryList)
	canonicalHeaders := strings.Join(result, "\n")
	canonicalRequest := c.method + "\n" + path + "\n" + query + "\n" + canonicalHeaders

	hs := hmac.New(sha256.New, []byte(signKey))
	hs.Write([]byte(canonicalRequest))
	signature := hex.EncodeToString(hs.Sum(nil))

	authorization := authStringPrefix + "/host;x-bce-date/" + signature

	ser := &http.Client{}
	requrl := fmt.Sprintf("http://%s%s?%s", c.host, path, query)
	req, err := http.NewRequest(c.method, requrl, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", authorization)
	resp, err := ser.Do(req)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	dataByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(dataByte, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *BCMClient) buildQueryParams(queryList []QueryParam) string {
	s := ""
	if len(queryList) == 0 {
		return s
	}
	for _, item := range queryList {
		s += fmt.Sprintf("&%s=%v", url.QueryEscape(item.K), url.QueryEscape(item.V))
	}
	return s[1:]
}

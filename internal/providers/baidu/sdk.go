package baidu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type BCMClient struct {
	ak   string
	sk   string
	host string
	http *http.Client
}

func NewBCMClient(AK, SK, host string) *BCMClient {
	return &BCMClient{
		ak:   AK,
		sk:   SK,
		host: host,
		http: &http.Client{},
	}
}

type QueryParam struct {
	K, V string
}

func (c *BCMClient) Send(path string, queryList []QueryParam) ([]byte, error) {
	timeStamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	authStringPrefix := fmt.Sprintf("bce-auth-v1/%s/%s/10000", c.ak, timeStamp)
	h := hmac.New(sha256.New, []byte(c.sk))
	h.Write([]byte(authStringPrefix))
	signKey := hex.EncodeToString(h.Sum(nil))

	query := c.buildQueryParams(queryList)
	canonicalRequest := "GET\n" + path + "\n" + query + "\nhost:" + c.host + "\nx-bce-date:" + url.QueryEscape(timeStamp)

	hs := hmac.New(sha256.New, []byte(signKey))
	hs.Write([]byte(canonicalRequest))
	signature := hex.EncodeToString(hs.Sum(nil))

	authorization := authStringPrefix + "/host;x-bce-date/" + signature

	ser := &http.Client{}
	reqUrl := fmt.Sprintf("https://%s%s?%s", c.host, path, query)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", c.host)
	req.Header.Set("x-bce-date", timeStamp)
	req.Header.Set("Authorization", authorization)
	resp, err := ser.Do(req)
	if err != nil {
		return nil, err
	}

	dataByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return dataByte, nil
}

func (c *BCMClient) buildQueryParams(queryList []QueryParam) string {
	s := ""
	if len(queryList) == 0 {
		return s
	}

	params := url.Values{}
	for _, item := range queryList {
		params.Add(item.K, item.V)
	}
	keys := make([]string, len(params))
	sort.Strings(keys)
	query := params.Encode()
	return query
}

package mixpanel

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	mixpanelApiBaseUrl = "http://api.mixpanel.com"
)

type Mixpanel struct {
	Token   string
	BaseUrl string
}

func NewMixpanel(token string) *Mixpanel {
	m := new(Mixpanel)
	m.Token = token
	m.BaseUrl = mixpanelApiBaseUrl
	return m
}

func (m *Mixpanel) makeRequest(method string, endpoint string, paramMap map[string]string) ([]byte, error) {
	var (
		err error
		req *http.Request
		b   io.Reader
	)

	if endpoint == "" {
		return []byte{}, errors.New("Endpoint missing")
	}

	endpoint = fmt.Sprintf("%v/%v", m.BaseUrl, endpoint)

	if paramMap == nil {
		paramMap = map[string]string{}
	}

	params := url.Values{}
	for k, v := range paramMap {
		params[k] = []string{v}
	}

	if method == "GET" {
		enc := params.Encode()
		if enc != "" {
			endpoint = endpoint + "?" + enc
		}
		req, err = http.NewRequest(method, endpoint, nil)
	} else if method == "POST" {
		b = strings.NewReader(params.Encode())
		req, err = http.NewRequest(method, endpoint, b)
	} else {
		err = fmt.Errorf("Method not supported: %v", method)
	}

	if err != nil {
		return []byte{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (m *Mixpanel) makeRequestWithData(method string, endpoint string, data map[string]interface{}) ([]byte, error) {
	var resp []byte

	json, err := json.Marshal(data)
	if err != nil {
		return resp, err
	}

	dataStr := base64.StdEncoding.EncodeToString(json)
	if err != nil {
		return resp, err
	}

	return m.makeRequest(method, endpoint, map[string]string{"data": dataStr})
}

func (m *Mixpanel) Track(distinctId string, event string, params map[string]string) error {
	if distinctId != "" {
		params["distinct_id"] = distinctId
	}
	params["token"] = m.Token
	params["mp_lib"] = "timehop/go-mixpanel"

	data := map[string]interface{}{"event": event, "properties": params}
	_, err := m.makeRequestWithData("GET", "track", data)
	if err != nil {
		return err
	}

	return nil
}

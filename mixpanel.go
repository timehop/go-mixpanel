package mixpanel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiBaseURL = "http://api.mixpanel.com"
	library    = "timehop/go-mixpanel"
)

// Mixpanel is a client to talk to the API
type Mixpanel struct {
	Token   string
	BaseUrl string
}

// Properties are key=value pairs that decorate an event or a profile.
type Properties map[string]interface{}

// Operation is an action performed on a user profile.
// Typically this is $set or $unset, but others are available.
type Operation struct {
	Name   string
	Values Properties
}

// NewMixpanel returns a configured client.
func NewMixpanel(token string) *Mixpanel {
	return &Mixpanel{
		Token:   token,
		BaseUrl: apiBaseURL,
	}
}

// Track sends event data with optional metadata.
func (m *Mixpanel) Track(distinctID string, event string, props Properties) error {
	if distinctID != "" {
		props["distinct_id"] = distinctID
	}
	props["token"] = m.Token
	props["mp_lib"] = library

	data := map[string]interface{}{"event": event, "properties": props}
	return m.makeRequestWithData("GET", "track", data)
}

// Engage updates profile data.
func (m *Mixpanel) Engage(distinctID string, props Properties, op *Operation) error {
	if distinctID != "" {
		props["$distinct_id"] = distinctID
	}
	props["$token"] = m.Token
	props["mp_lib"] = library
	if op.Name == "$unset" {
		keys := []interface{}{}
		for key, _ := range op.Values {
			keys = append(keys, key)
		}
		props[op.Name] = keys
	} else {
		props[op.Name] = op.Values
	}

	return m.makeRequestWithData("GET", "engage", props)
}

// RedirectURL returns a url that, when clicked, will track the given data and then redirect to provided url.
func (m *Mixpanel) RedirectURL(distinctId, event, uri string, props Properties) (string, error) {
	if distinctId != "" {
		props["$distinct_id"] = distinctId
	}
	props["$token"] = m.Token
	props["mp_lib"] = library

	data := map[string]interface{}{"event": event, "properties": props}
	json, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"data":     base64.StdEncoding.EncodeToString(json),
		"redirect": uri,
	}
	query := url.Values{}
	for k, v := range params {
		query[k] = []string{v}
	}
	return fmt.Sprintf("%s/%s?%s", m.BaseUrl, "track", query.Encode()), nil
}

func (m *Mixpanel) makeRequest(method string, endpoint string, paramMap map[string]string) error {
	var (
		err error
		req *http.Request
		r   io.Reader
	)

	if endpoint == "" {
		return fmt.Errorf("endpoint missing")
	}

	endpoint = fmt.Sprintf("%s/%s", m.BaseUrl, endpoint)

	if paramMap == nil {
		paramMap = map[string]string{}
	}

	params := url.Values{}
	for k, v := range paramMap {
		params[k] = []string{v}
	}

	switch method {
	case "GET":
		enc := params.Encode()
		if enc != "" {
			endpoint = endpoint + "?" + enc
		}
	case "POST":
		r = strings.NewReader(params.Encode())
	default:
		return fmt.Errorf("method not supported: %v", method)
	}

	req, err = http.NewRequest(method, endpoint, r)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// The API documentation states that success will be reported with either "1" or "1\n".
	if strings.Trim(string(b), "\n") != "1" {
		return fmt.Errorf("request failed - %s", b)
	}
	return nil
}

func (m *Mixpanel) makeRequestWithData(method string, endpoint string, data Properties) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataStr := base64.StdEncoding.EncodeToString(b)
	return m.makeRequest(method, endpoint, map[string]string{"data": dataStr})
}

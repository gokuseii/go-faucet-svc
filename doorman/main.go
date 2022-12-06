package doorman

import (
	"bytes"
	"encoding/json"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Connector struct {
	ServiceUrl string
	Client     *http.Client
}

func NewConnector(serviceUrl string) Connector {
	return Connector{
		ServiceUrl: serviceUrl,
		Client: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

func (c *Connector) Authenticate(r *http.Request) (*resources.User, error) {
	endpoint, err := url.Parse(c.ServiceUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create authentication request out of incoming request")
	}

	req, err := createAuthenticationRequest(r, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create authentication request out of incoming request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errors.New("failed to authenticate request due to invalid authentication info")
	case http.StatusUnauthorized:
		return nil, errors.New("requester is unauthorized")
	case http.StatusInternalServerError:
		return nil, errors.New("internal server error occurred while authentication")
	}

	var user resources.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(err, "failed to decode user response")
	}

	r.Header.Set("User-Id", user.Id)
	return &user, nil
}

func createAuthenticationRequest(r *http.Request, endpoint *url.URL) (*http.Request, error) {
	drmUrl, err := endpoint.Parse("/doorman/authenticate")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse doorman service url")
	}

	rBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the original body")
	}

	r.Body = io.NopCloser(bytes.NewReader(rBody))
	req, err := http.NewRequest("GET", drmUrl.String(), bytes.NewReader(rBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create authentication request")
	}

	authHeader, authHeaderOk := GetHeader(r, "Authorization")
	if authHeaderOk {
		req.Header.Set("Authorization", authHeader)
	}

	return req, nil
}

func GetHeader(r *http.Request, headerName string) (headerValue string, ok bool) {
	headerValue = r.Header.Get(headerName)
	ok = headerValue != ""
	return
}

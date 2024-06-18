package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"errors"
)

const (
	restUrl         = "https://api.kraken.com/"
	restVersion     = "0"
	restPrivatePath = restVersion + "/private"
	restPrivateUrl  = restUrl + restPrivatePath
	restPublicUrl   = restUrl + restVersion + "/public"
)

type getWebSocketTokenResp struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

type krakenResponse struct {
	Error  []string    `json:"error"`
	Result interface{} `json:"result"`
}

func (priv *private) request(method string, data url.Values, retType interface{}) error {
	req, err := priv.prepareRequest(method, data)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during request execution: %w", err)
	}
	defer resp.Body.Close()
	err = priv.parseResponse(resp, retType)
	if err != nil {
		return fmt.Errorf("error during response parsing: %w", err)
	}
	return nil
}

func (priv *private) prepareRequest(method string, data url.Values) (*http.Request, error) {
	if data == nil {
		data = url.Values{}
	}
	requestURL := ""
	requestURL = fmt.Sprintf("%s/%s", restPrivateUrl, method)
	data.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error during request creation: %w", err)
	}

	urlPath := fmt.Sprintf("/%s/%s", restPrivatePath, method)
	req.Header.Add("API-Key", priv.pubkey)
	signature, err := priv.getSign(urlPath, data)
	if err != nil {
		return nil, fmt.Errorf("invalid secret key: %w", err)
	}
	req.Header.Add("API-Sign", signature)

	return req, nil
}

func (priv *private) getSign(requestURL string, data url.Values) (string, error) {
	sha := sha256.New()

	if _, err := sha.Write([]byte(data.Get("nonce") + data.Encode())); err != nil {
		return "", err
	}
	hashData := sha.Sum(nil)
	s, err := base64.StdEncoding.DecodeString(priv.privkey)
	if err != nil {
		return "", err
	}
	hmacObj := hmac.New(sha512.New, s)

	if _, err := hmacObj.Write(append([]byte(requestURL), hashData...)); err != nil {
		return "", err
	}
	hmacData := hmacObj.Sum(nil)
	return base64.StdEncoding.EncodeToString(hmacData), nil
}

func (priv *private) parseResponse(response *http.Response, retType interface{}) error {
	if response.StatusCode != 200 {
		return fmt.Errorf("invalid status code %d", response.StatusCode)
	}
	if response.Body == nil {
		return errors.New("no response body")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("can not read error body: %w", err)
	}

	var retData krakenResponse
	if retType != nil {
		retData.Result = retType
	}

	if err = json.Unmarshal(body, &retData); err != nil {
		return fmt.Errorf("can't unmarshal response: %w", err)
	}

	if len(retData.Error) > 0 {
		return errors.New("kraken returned errors: " + strings.Join(retData.Error, ","))
	}

	return nil
}

func (pr *private) GetWebSocketsToken() (response getWebSocketTokenResp, err error) {
	err = pr.request("GetWebSocketsToken", nil, &response)
	return
}

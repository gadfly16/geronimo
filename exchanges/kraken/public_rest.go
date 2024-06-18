package kraken

// import (
// 	"fmt"
// 	"net/http"
// 	"net/url"
// 	"strings"
// )

// func (pub *public) prepareRequest(method string, data url.Values) (*http.Request, error) {
// 	if data == nil {
// 		data = url.Values{}
// 	}
// 	requestURL := ""
// 	requestURL = fmt.Sprintf("%s/%s", APIPublicUrl, method)
// 	req, err := http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
// 	if err != nil {
// 		return nil, fmt.Errorf("error during request creation: %w", err)
// 	}
// 	return req, nil
// }

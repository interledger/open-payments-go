// These methods are currently broken. They depend on signing headers to authorize request.

package openpayments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
)

type GrantControllers struct{
	httpClient *http.Client
}

type Grant struct {
	AccessToken as.AccessToken
	Continue as.Continue
}

func (g *GrantControllers) Request(url string, requestBody as.PostRequestJSONBody) (Grant, error) {
		reqBodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return Grant{}, fmt.Errorf("failed to marshal request body: %s", err)
		}

    resp, err := g.httpClient.Post(url,  "application/json", bytes.NewBuffer(reqBodyBytes))
    if err != nil {
        return Grant{}, err
    }
    defer resp.Body.Close()

		// TODO: remove this debug log after authorization implemented and this is confirmed working 
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Response Body:", string(body))

    if resp.StatusCode != http.StatusOK {
        return Grant{}, fmt.Errorf("failed to perform grant request: %s", resp.Status)
    }

    var grantResponse Grant
    err = json.NewDecoder(resp.Body).Decode(&grantResponse)
    if err != nil {
        return Grant{}, fmt.Errorf("failed to decoding response body: %s", err)
    }

    return grantResponse, nil
}
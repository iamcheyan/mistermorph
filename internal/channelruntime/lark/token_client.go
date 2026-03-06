package lark

import (
	"net/http"

	larkapi "github.com/quailyquaily/mistermorph/internal/larkapi"
)

const (
	defaultLarkBaseURL        = larkapi.DefaultBaseURL
	defaultTokenRefreshBefore = larkapi.DefaultRefreshBefore
	tenantAccessTokenAPIPath  = larkapi.TenantAccessTokenPath
)

type tenantTokenClient = larkapi.TenantTokenClient
type TenantTokenClient = larkapi.TenantTokenClient

func newTenantTokenClient(httpClient *http.Client, baseURL, appID, appSecret string) *tenantTokenClient {
	return larkapi.NewTenantTokenClient(httpClient, baseURL, appID, appSecret)
}

func NewTenantTokenClient(httpClient *http.Client, baseURL, appID, appSecret string) *TenantTokenClient {
	return larkapi.NewTenantTokenClient(httpClient, baseURL, appID, appSecret)
}

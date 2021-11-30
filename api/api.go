package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	g_self_http_client = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			TLSHandshakeTimeout:   10 * time.Second,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
	}
)

func SendRequest(ctx context.Context, req *http.Request) (errCode int, body []byte, err error) {
	UUID := uuid.NewV4()
	client := g_self_http_client
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("x-message-id", UUID.String())
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	errCode = resp.StatusCode
	body, err = ioutil.ReadAll(resp.Body)
	return
}

type (
	Login struct {
		AccountType string `json:"account_type"`
		GrantType   string `json:"grant_type"`
		Email       string `json:"email"`
		Password    string `json:"password"`
	}
	LoginResponse struct {
		OptStatus   string `json:"OPT_STATUS"`
		Data        Data   `json:"DATA"`
		OptStatusCh string `json:"OPT_STATUS_CH"`
		OptStatusEN string `json:"OPT_STATUS_EN"`
	}
	Data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expire       int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Email        string `json:"email"`
	}
)

func (c *Client) AuthToken() (err error) {
	if c.AuthLogin == nil {
		err = errors.New("authlogin is nil")
		return
	}
	msg, err := json.Marshal(c.AuthLogin)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%v/auth/login", c.RemoteHost)
	req, err := http.NewRequest("POST", url, bytes.NewReader(msg))
	if err != nil {
		return
	}
	_, body, err := SendRequest(context.TODO(), req)
	if err != nil {
		return
	}
	loginResp := &LoginResponse{}
	err = json.Unmarshal(body, loginResp)
	if err != nil {
		return
	}
	if !strings.EqualFold(loginResp.OptStatus, "SUCCESS") || loginResp.Data.AccessToken == "" {
		err = fmt.Errorf("login optstatus or token invalid, status=%v, accesstoken=%v", loginResp.OptStatus, loginResp.Data.AccessToken)
		return
	}
	c.token = loginResp.Data.AccessToken
	c.refreshToken = loginResp.Data.RefreshToken
	fmt.Printf("authtoken, token=%v, refreshtoken=%v\n", c.token, c.refreshToken)
	return nil
}

type (
	Refresh struct {
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
	}
	RefreshResponse struct {
		OptStatus     string      `json:"OPT_STATUS"`
		Data          RefreshData `json:"DATA"`
		OPT_STATUS_CH string      `json:"OPT_STATUS_CH"`
		OPT_STATUS_EN string      `json:"OPT_STATUS_EN"`
	}
	RefreshData struct {
		AccessToken       string `json:"access_token"`
		RefreshTokenValue string `json:"refresh_token"`
		Expire            int    `json:"expires_in"`
		TokenType         string `json:"token_type"`
	}
)

func (c *Client) RefreshToken() (err error) {
	if c.token == "" {
		return c.AuthToken()
	}
	msg, err := json.Marshal(&Refresh{
		GrantType:    "refresh_token",
		RefreshToken: c.refreshToken,
	})
	if err != nil {
		return
	}
	url := fmt.Sprintf("%v/auth/refresh_token", c.RemoteHost)
	req, err := http.NewRequest("POST", url, bytes.NewReader(msg))
	if err != nil {
		return
	}
	_, body, err := SendRequest(context.TODO(), req)
	if err != nil {
		return
	}
	ref := &RefreshResponse{}
	err = json.Unmarshal(body, ref)
	if err != nil {
		return
	}
	if !strings.EqualFold(ref.OptStatus, "SUCCESS") || ref.Data.AccessToken == "" {
		err = fmt.Errorf("auth optstatus or token invalid, status=%v, accesstoken=%v", ref.OptStatus, ref.Data.AccessToken)
		return
	}
	c.token = ref.Data.AccessToken
	c.refreshToken = ref.Data.RefreshTokenValue
	return nil
}

var httpClient = new(Client)

func SetClient(c *Client) {
	httpClient = c
}
func PutWgvpnResource(routes []string) error {
	return httpClient.PutWgvpnResource(routes)
}

type Client struct {
	AuthLogin    *Login
	RemoteHost   string
	token        string
	refreshToken string
	MaxTry       int
}
type (
	InstanceResponse struct {
		Data        []WanAccessInstanceResponse `json:"DATA,omitempty"`
		Desctiption string                      `json:"DESCRIPTION,omitempty"`
		OptStatus   string                      `json:"OPT_STATUS,omitempty"`
	}
	WanAccessInstanceResponse struct {
		AuthType           int                     `json:"auth_type"`
		Created            string                  `json:"created_at"`
		Desctiption        string                  `json:"description"`
		Name               string                  `json:"name"`
		Protocol           int                     `json:"protocol,omitempty"`
		RadiusServerIP     string                  `json:"radius_server_ip,omitempty"`
		RadiusServerPasswd string                  `json:"radius_server_passwd,omitempty"`
		RadiusServerPort   int                     `json:"radius_server_port,omitempty"`
		ScriptName         string                  `json:"script_name,omitempty"`
		Status             string                  `json:"status,omitempty"`
		UserUUID           string                  `json:"user_uuid"`
		UUID               string                  `json:"uuid"`
		XRounters          []AccessXRouterResponse `json:"XRounters"`
	}
	AccessXRouterResponse struct {
		MgmtIP string `json:"mgmt_ip"`
		Name   string `json:"name"`
	}
)

func (c *Client) GetWanAccessInstances() (insp InstanceResponse, err error) {

	if c.token == "" {
		if err = c.AuthToken(); err != nil {
			return
		}
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/wan-service/v1/wan-access/instance", c.RemoteHost), nil)
	if err != nil {
		return
	}
	try := 0
Loop:
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
	statusCode, body, err := SendRequest(context.TODO(), req)
	if err != nil {
		return
	}
	if statusCode != 200 {
		if try <= c.MaxTry {
			if err = c.RefreshToken(); err != nil {
				return
			}
			try++
			goto Loop
		} else {
			err = fmt.Errorf("GetWanAccessInstances error, statuscode=%v, token=%v", statusCode, c.token)
			return
		}
	}
	var rep = &InstanceResponse{}
	err = json.Unmarshal(body, rep)
	if err != nil {
		return
	}
	return *rep, nil
}

//

type (
	WgvpnResponse struct {
		Data        []CpeWgVpnResponse `json:"DATA,omitempty"`
		Description string             `json:"DESCRIPTION,omitempty"`
		OptStatus   string             `json:"OPT_STATUS,omitempty"`
	}
	CpeWgVpnResponse struct {
		CpeUUID    string   `json:"cpe_uuid,omitempty"`
		DnsServer  []string `json:"dns_server,omitempty"`
		IPPoolUUID string   `json:"ip_pool_uuid"`
		Name       string   `json:"name"`
		Port       int      `json:"port"`
		Routes     []string `json:"routes"`
		Status     string   `json:"status,omitempty"`
		UUID       string   `json:"uuid"`
	}
)

func (c *Client) GetWgvpnResource() (wr WgvpnResponse, uuid string, err error) {
	if c.token == "" {
		if err = c.AuthToken(); err != nil {
			return
		}
	}
	insp, err := c.GetWanAccessInstances()
	if err != nil {
		return
	}
	for _, data := range insp.Data {
		if data.Name == "yunshan" {
			uuid = data.UUID
			break
		}
	}
	if uuid == "" {
		err = fmt.Errorf("GetWgvpnResource:can not find uuid of yunshan")
		return
	}
	url := fmt.Sprintf("%v/api/wan-service/v1/wan-access/instance/%v/wgvpn-resource", c.RemoteHost, uuid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	try := 0
Loop:
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
	statusCode, body, err := SendRequest(context.TODO(), req)
	if err != nil {
		return
	}
	if statusCode != 200 {
		if try <= c.MaxTry {
			if err = c.RefreshToken(); err != nil {
				return
			}
			try++
			goto Loop
		} else {
			err = fmt.Errorf("GetWgvpnResource error, statuscode=%v, token=%v, refreshtoken=%v", statusCode, c.token, c.refreshToken)
			return
		}

	}
	var wgvpnResponse = &WgvpnResponse{}
	err = json.Unmarshal(body, wgvpnResponse)
	if err != nil {
		return
	}
	return *wgvpnResponse, uuid, nil
}

//
type (
	WgvpnReq struct {
		Delta        CpeWgVpnResource `json:"delta"`
		UUID         string           `json:"uuid"`
		ResourceUUID string           `json:"resource_uuid"`
	}
	CpeWgVpnResource struct {
		CpeUUID    string   `json:"cpe_uuid,omitempty"`
		DnsServer  []string `json:"dns_server,omitempty"`
		IPPoolUUID string   `json:"ip_pool_uuid"`
		Name       string   `json:"name"`
		Port       int      `json:"port"`
		Routes     []string `json:"routes"`
	}
)

func (c *Client) PutWgvpnResource(routes []string) (err error) {
	if c.token == "" {
		if err = c.AuthToken(); err != nil {
			return
		}
	}
	wr, uuid, err := c.GetWgvpnResource()
	if err != nil {
		return
	}
	var data CpeWgVpnResponse
	for _, v := range wr.Data {
		if v.Name == "vpngw" {
			data = v
			break
		}
	}
	if data.UUID == "" {
		return fmt.Errorf("can not find resouece_uuid of vpngw")
	}
	wgVpnReq := &WgvpnReq{
		Delta: CpeWgVpnResource{
			CpeUUID:    data.CpeUUID,
			DnsServer:  data.DnsServer,
			IPPoolUUID: data.IPPoolUUID,
			Name:       data.Name,
			Port:       data.Port,
			Routes:     routes,
		},
		UUID:         uuid,
		ResourceUUID: data.UUID,
	}
	wgVpnReqByte, err := json.Marshal(wgVpnReq)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%v/api/wan-service/v1/wan-access/instance/%v/wgvpn-resource/%v", c.RemoteHost, wgVpnReq.UUID, wgVpnReq.ResourceUUID)
	fmt.Println("wgVpnReqByte=", string(wgVpnReqByte))
	req, err := http.NewRequest("PUT", url, bytes.NewReader(wgVpnReqByte))
	if err != nil {
		return
	}
	try := 0
Loop:
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
	statusCode, _, err := SendRequest(context.TODO(), req)
	if statusCode != 200 {
		if try <= c.MaxTry {
			if err = c.RefreshToken(); err != nil {
				return err
			}
			try++
			goto Loop
		} else {
			err = fmt.Errorf("PutWgvpnResource error, statuscode=%v, token=%v", statusCode, c.token)
		}

	}
	return
}

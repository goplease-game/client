package contentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goplease-game/client/config"
	"github.com/goplease-game/content-server/app/ds"
	"github.com/shirou/gopsutil/v3/host"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	baseURL string
	cl      *http.Client
}

func New() *Client {
	return &Client{
		baseURL: serverURL(),
		cl: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

type DeviceAuthRequest struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

// CreateDeviceAuthRequest starts a new device authorization flow, returning
// the codes the client needs to poll and the URL to open in a browser.
func (c *Client) CreateDeviceAuthRequest(ctx context.Context) (*DeviceAuthRequest, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/api/device-auth/code/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create device auth request: unexpected status %d", resp.StatusCode)
	}

	var out DeviceAuthRequest
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

type ConsumeDeviceAuthResult struct {
	Status    DeviceAuthRequestStatus `json:"status"`
	User      *ds.User                `json:"user,omitempty"`
	AuthToken *string                 `json:"auth_token,omitempty"` //nolint:gosec
}

// ConsumeDeviceAuthRequest polls the server once with the given device_code.
func (c *Client) ConsumeDeviceAuthRequest(ctx context.Context, deviceCode string) (*ConsumeDeviceAuthResult, error) {
	body, _ := json.Marshal(map[string]string{"code": deviceCode})

	req, err := c.newRequest(ctx, http.MethodPost, "/api/device-auth/consume/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consume device auth request: unexpected status %d", resp.StatusCode)
	}

	var out ConsumeDeviceAuthResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// userAgent builds the User-Agent header for native clients: a
// human-readable string, ready to display as-is on the device-auth
// confirmation page. The server recognizes it by the "goplease-client"
// prefix and passes it through unparsed.
func userAgent() string {
	return fmt.Sprintf("goplease-client on %s", clientInfo())
}

func clientInfo() string {
	info, err := host.Info()
	if err != nil {
		return "unknown device"
	}

	return fmt.Sprintf("%s at %s", info.Platform, info.Hostname)
}

func serverURL() string {
	conf := config.Get()
	scheme := "http"
	if conf.ContentServer.Secure {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, conf.ContentServer.Addr)
}

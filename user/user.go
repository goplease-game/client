package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/goplease-game/client/content-client"
)

type Status int

const (
	StatusGuest Status = iota
	StatusPendingAuth
	StatusAuthenticated
)

const (
	pollInterval    = 5 * time.Second
	deviceAuthTTL   = 10 * time.Minute
	maxPollInterval = 30 * time.Second
)

// User holds client-side auth state: guest, mid-login, or authenticated.
// Safe for concurrent use — Status()/Username() are read every frame from
// the UI goroutine while StartLogin() polls on its own goroutine.
type User struct {
	mu       sync.RWMutex
	status   Status
	username string
	token    string

	cl         *contentclient.Client
	cancelPoll context.CancelFunc

	// OnAuthError is called on the polling goroutine if login fails or is
	// denied. UI code should set this to show an error and reset to Guest.
	OnAuthError func(err error)

	// OnVerificationURI is called with the browser URL the user should open
	// (e.g. via game.OpenURL). Also receives the human-readable user_code as
	// a fallback for manual entry.
	OnVerificationURI func(uri, userCode string)
}

// Load restores a persisted session token, if any, and returns a User in
// the Authenticated or Guest state accordingly.
func Load(cl *contentclient.Client) *User {
	u := &User{cl: cl, status: StatusGuest}

	token, username, ok := loadPersistedSession()
	if ok {
		u.token = token
		u.username = username
		u.status = StatusAuthenticated
	}

	return u
}

func (u *User) Status() Status {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.status
}

func (u *User) Username() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.username
}

func (u *User) Token() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.token
}

// StartLogin begins the device authorization flow: requests a code pair,
// notifies the caller of the verification URL via OnVerificationURI, and
// starts polling in the background. Safe to call again to retry after a
// failure (cancels any in-flight poll first).
func (u *User) StartLogin() {
	u.mu.Lock()
	if u.cancelPoll != nil {
		u.cancelPoll()
	}
	u.status = StatusPendingAuth
	u.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), deviceAuthTTL)

	u.mu.Lock()
	u.cancelPoll = cancel
	u.mu.Unlock()

	go u.runLogin(ctx)
}

// CancelLogin aborts an in-progress login and returns to the Guest state.
func (u *User) CancelLogin() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.cancelPoll != nil {
		u.cancelPoll()
		u.cancelPoll = nil
	}
	if u.status == StatusPendingAuth {
		u.status = StatusGuest
	}
}

func (u *User) runLogin(ctx context.Context) {
	authReq, err := u.cl.CreateDeviceAuthRequest(ctx)
	if err != nil {
		u.failLogin(fmt.Errorf("start device auth: %w", err))
		return
	}

	if u.OnVerificationURI != nil {
		u.OnVerificationURI(authReq.VerificationURI, authReq.UserCode)
	}

	interval := pollInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			u.failLogin(fmt.Errorf("device auth timed out or was cancelled"))
			return
		case <-ticker.C:
		}

		res, err := u.cl.ConsumeDeviceAuthRequest(ctx, authReq.DeviceCode)
		if err != nil {
			// Transient network error — keep polling rather than aborting
			// the whole login on one failed request.
			continue
		}

		switch res.Status {
		case "pending":
			continue
		case "approved":
			if res.AuthToken == nil {
				u.failLogin(fmt.Errorf("server approved request but returned no token"))
				return
			}
			if res.User == nil {
				u.failLogin(fmt.Errorf("server approved request but returned no user"))
				return
			}
			u.completeLogin(*res.AuthToken, res.User.Username)
			return
		case "denied":
			u.failLogin(fmt.Errorf("login was denied"))
			return
		case "expired":
			u.failLogin(fmt.Errorf("login code expired"))
			return
		case "slow_down":
			interval = min(interval*2, maxPollInterval)
			ticker.Reset(interval)
		default:
			u.failLogin(fmt.Errorf("unexpected device auth status: %s", res.Status))
			return
		}
	}
}

func (u *User) completeLogin(token, username string) {
	persistSession(token, username)

	u.mu.Lock()
	u.token = token
	u.username = username
	u.status = StatusAuthenticated
	u.cancelPoll = nil
	u.mu.Unlock()
}

func (u *User) failLogin(err error) {
	u.mu.Lock()
	u.status = StatusGuest
	u.cancelPoll = nil
	u.mu.Unlock()

	if u.OnAuthError != nil {
		u.OnAuthError(err)
	}
}

// Logout clears the session, both in memory and in persisted storage.
func (u *User) Logout() {
	u.mu.Lock()
	u.token = ""
	u.username = ""
	u.status = StatusGuest
	u.mu.Unlock()

	clearPersistedSession()
}

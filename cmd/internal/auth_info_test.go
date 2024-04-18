package internal

import (
	"strings"
	"testing"
)

func TestNewBasicAuthInfo(t *testing.T) {
	// Create new BasicAuthInfo struct.
	authInfo := NewBasicAuthInfo("foo", "bar")
	if authInfo.Username != "foo" {
		t.Errorf("invalid username: expected=%q  actual=%q", "foo", authInfo.Username)
	}
	if authInfo.Password != "bar" {
		t.Errorf("invalid password: expected=%q  actual=%q", "bar", authInfo.Password)
	}
}

func TestNewOAuthToken(t *testing.T) {
	// Create new OAuthToken struct.
	token := NewOAuthToken("foo")
	if token.Token != "foo" {
		t.Errorf("invalid token: expected=%q  actual=%q", "foo", token.Token)
	}
}

func TestNewPrivateToken(t *testing.T) {
	// Create new PrivateToken struct.
	token := NewPrivateToken("foo")
	if token.Token != "foo" {
		t.Errorf("invalid token: expected=%q  actual=%q", "foo", token.Token)
	}
}

func TestNewBasicAuthInfoFromJSON(t *testing.T) {
	type Data []struct {
		root string
		username string
		password string
		err error
	}

	data := Data{
		{
			root: `{"username": "foo", "password": "bar"}`,
			username: "foo",
			password: "bar",
			err: nil,
		},
		{
			root: `{"oauth-token": "<token>"}`,
			username: "",
			password: "",
			err: ErrAuthInfoInvalidJSON,
		},
		{
			root: `{"private-token": "<token>"}`,
			username: "",
			password: "",
			err: ErrAuthInfoInvalidJSON,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		authInfo, err := NewBasicAuthInfoFromJSON(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if authInfo.Username != d.username {
				t.Errorf("invalid username: expected=%q  actual=%q", "<username>", authInfo.Username)
			}
			if authInfo.Password != d.password {
				t.Errorf("invalid password: expected=%q  actual=%q", "<password>", authInfo.Password)
			}
		}
	}
}

func TestNewOAuthTokenFromJSON(t *testing.T) {
	type Data []struct {
		root string
		token string
		err error
	}

	data := Data{
		{
			root: `{"oauth-token": "<token>"}`,
			token: "<token>",
			err: nil,
		},
		{
			root: `{"private-token": "<token>"}`,
			token: "",
			err: ErrAuthInfoInvalidJSON,
		},
		{
			root: `{"username": "foo", "password": "bar"}`,
			token: "",
			err: ErrAuthInfoInvalidJSON,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		token, err := NewOAuthTokenFromJSON(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if token.Token != d.token {
				t.Errorf("invalid token: expected=%q  actual=%q", "<token>", token.Token)
			}
		}
	}
}

func TestPrivateTokenFromJSON(t *testing.T) {
	type Data []struct {
		root string
		token string
		err error
	}

	data := Data{
		{
			root: `{"private-token": "<token>"}`,
			token: "<token>",
			err: nil,
		},
		{
			root: `{"oauth-token": "<token>"}`,
			token: "",
			err: ErrAuthInfoInvalidJSON,
		},
		{
			root: `{"username": "foo", "password": "bar"}`,
			token: "",
			err: ErrAuthInfoInvalidJSON,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		token, err := NewPrivateTokenFromJSON(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if token.Token != d.token {
				t.Errorf("invalid token: expected=%q  actual=%q", "<token>", token.Token)
			}
		}
	}
}

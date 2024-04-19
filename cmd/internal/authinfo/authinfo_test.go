package authinfo

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

func TestNewBasicAuthInfoFromXML(t *testing.T) {
	type Data []struct {
		root string
		username string
		password string
		err error
	}

	data := Data{
		{
			root: `
                <AuthInfo>
                    <username>foo</username>
                    <password>bar</password>
                </AuthInfo>`,
			username: "foo",
			password: "bar",
			err: nil,
		},
		{
			root: `
                <AuthInfo>
                    <oauth-token>token</oauth-token>
                </AuthInfo>`,
			username: "",
			password: "",
			err: ErrAuthInfoInvalidXML,
		},
		{
			root: `
                <AuthInfo>
                    <private-token>token</private-token>
                </AuthInfo>`,
			username: "",
			password: "",
			err: ErrAuthInfoInvalidXML,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		authInfo, err := NewBasicAuthInfoFromXML(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if authInfo.Username != d.username {
				t.Errorf("invalid username: expected=%q  actual=%q", d.username, authInfo.Username)
			}
			if authInfo.Password != d.password {
				t.Errorf("invalid password: expected=%q  actual=%q", d.password, authInfo.Password)
			}
		}
	}
}

func TestNewOAuthTokenFromXML(t *testing.T) {
	type Data []struct {
		root string
		token string
		err error
	}

	data := Data{
		{
			root: `
                <AuthInfo>
                    <oauth-token>token</oauth-token>
                </AuthInfo>`,
			token: "token",
			err: nil,
		},
		{
			root: `
                <AuthInfo>
                    <username>foo</username>
                    <password>bar</password>
                </AuthInfo>`,
			err: ErrAuthInfoInvalidXML,
		},
		{
			root: `
                <AuthInfo>
                    <private-token>token</private-token>
                </AuthInfo>`,
			token: "token",
			err: ErrAuthInfoInvalidXML,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		token, err := NewOAuthTokenFromXML(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if token.Token != d.token {
				t.Errorf("invalid token: expected=%q  actual=%q", d.token, token.Token)
			}
		}
	}
}

func TestPrivateTokenFromXML(t *testing.T) {
	type Data []struct {
		root string
		token string
		err error
	}

	data := Data{
		{
			root: `
                <AuthInfo>
                    <private-token>token</private-token>
                </AuthInfo>`,
			token: "token",
			err: nil,
		},
		{
			root: `
                <AuthInfo>
                    <oauth-token>token</oauth-token>
                </AuthInfo>`,
			token: "token",
			err: ErrAuthInfoInvalidXML,
		},
		{
			root: `
                <AuthInfo>
                    <username>foo</username>
                    <password>bar</password>
                </AuthInfo>`,
			err: ErrAuthInfoInvalidXML,
		},
	}

	for _, d := range data {

		r := strings.NewReader(d.root)
		token, err := NewPrivateTokenFromXML(r)
		if err != d.err {	
			t.Fatalf("unexpected error: %v: %s", err, d.root)
		}
		if d.err == nil {
			if token.Token != d.token {
				t.Errorf("invalid token: expected=%q  actual=%q", d.token, token.Token)
			}
		}
	}
}

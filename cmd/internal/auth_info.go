// This file is used for reading the user's access token from a JSON
// file.  The file should have one of the following formats:
//
//     // Basic Authentication
//     {
//         "username": "<username>",
//         "password": "password"
//     }
//
//     // OAuth Token
//     {
//         "oauth-token: "token"
//     }
//
//     // Personal or Private Access Token
//     {
//         "private-token": "<token>"
//     }

package internal

import (
	"errors"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Errors
////////////////////////////////////////////////////////////////////////

var (
	ErrAuthInfoInvalidJSON = errors.New("invalid JSON")
)

////////////////////////////////////////////////////////////////////////
// AuthInfo
////////////////////////////////////////////////////////////////////////

// Interface implemented by types that return information used for
// authentication.
type AuthInfo interface {

	// CreateGitlabClient returns a new Gitlab Client based on the
	// authentication information provided by the user.  The options
	// parameter is the same "options" parameter that is passed into
	// the gitlab.New*Client() methods which can be used to tailor the
	// client for the user's purpose.
	CreateGitlabClient(options ...gitlab.ClientOptionFunc) (*gitlab.Client, error)
}

////////////////////////////////////////////////////////////////////////
// BasicAuthInfo
////////////////////////////////////////////////////////////////////////

// BasicAuthInfo holds username and password used for HTTP basic authentication.
type BasicAuthInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewBasicAuthInfo creates a new set of authentication information
// for HTTP basic authentication.
func NewBasicAuthInfo(username, password string) BasicAuthInfo {
	return BasicAuthInfo{
		Username: username,
		Password: password,
	}
}

// NewBasicAuthInfoFromJSON creates a new set of authentication
// information for HTTP basic authentication from the JSON accessible
// through the io.Reader.  The format of the JSON is as follows:
//
//  {
//      "username": "<username>",
//      "password": "password"
//  }
func NewBasicAuthInfoFromJSON(r io.Reader) (BasicAuthInfo, error) {
	result := BasicAuthInfo{}
	err := json.NewDecoder(r).Decode(&result)
	if err != nil {
		return BasicAuthInfo{}, err
	}
	if (len(result.Username) == 0) || (len(result.Password) == 0) {
		return BasicAuthInfo{}, ErrAuthInfoInvalidJSON
	}
	return result, nil
}

// CreateGitlabClient returns a new Gitlab Client that uses HTTP basic
// authentication.  The options parameter is the same "options"
// parameter that is passed into the gitlab.New*Client() methods which
// can be used to tailor the client for the user's purpose.
func (authInfo *BasicAuthInfo) CreateGitlabClient(options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	return gitlab.NewBasicAuthClient(
		authInfo.Username,
		authInfo.Password,
		options...)
}

////////////////////////////////////////////////////////////////////////
// OAuthToken
////////////////////////////////////////////////////////////////////////

// OAuthToken holds an OAuth access token.
type OAuthToken struct {
	Token string `json:"oauth-token"`
}

// NewOAuthToken creates a new set of authentication information for
// OAuth authentication.
func NewOAuthToken(token string) OAuthToken {
	return OAuthToken{
		Token: token,
	}
}

// NewOAuthTokenFromJSON creates a new set of authentication
// information for OAuth authentication from the JSON accessible
// through the io.Reader.  The format of the JSON is as follows:
//
//  {
//      "oauth-token": "<token>"
//  }
func NewOAuthTokenFromJSON(r io.Reader) (OAuthToken, error) {
	result := OAuthToken{}
	err := json.NewDecoder(r).Decode(&result)
	if err != nil {
		return OAuthToken{}, err
	}
	if len(result.Token) == 0 {
		return OAuthToken{}, ErrAuthInfoInvalidJSON
	}
	return result, nil
}

// CreateGitlabClient returns a new Gitlab Client that uses OAuth
// authentication.  The options parameter is the same "options"
// parameter that is passed into the gitlab.New*Client() methods which
// can be used to tailor the client for the user's purpose.
func (token *OAuthToken) CreateGitlabClient(options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	return gitlab.NewOAuthClient(token.Token, options...)
}

////////////////////////////////////////////////////////////////////////
// PrivateToken
////////////////////////////////////////////////////////////////////////

// PrivateToken holds a private or personal access token.
type PrivateToken struct {
	Token string `json:"private-token"`
}

// NewPrivateToken creates a new set of authentication information for
// private token or personal token authentication.
func NewPrivateToken(token string) PrivateToken {
	return PrivateToken{
		Token: token,
	}
}

// NewPrivateTokenFromJSON creates a new set of authentication
// information for private token or personal token authentication from
// the JSON accessible through the io.Reader.  The format of the JSON
// is as follows:
//
//  {
//      "private-token": "<token>"
//  }
func NewPrivateTokenFromJSON(r io.Reader) (PrivateToken, error) {
	result := PrivateToken{}
	err := json.NewDecoder(r).Decode(&result)
	if err != nil {
		return PrivateToken{}, err
	}
	if len(result.Token) == 0 {
		return PrivateToken{}, ErrAuthInfoInvalidJSON
	}
	return result, nil
}

// CreateGitlabClient returns a new Gitlab Client that uses private
// token or personal token authentication.  The options parameter is
// the same "options" parameter that is passed into the
// gitlab.New*Client() methods which can be used to tailor the client
// for the user's purpose.
func (token *PrivateToken) CreateGitlabClient(options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
	return gitlab.NewClient(token.Token, options...)
}

////////////////////////////////////////////////////////////////////////
// LoadAuthInfo()
////////////////////////////////////////////////////////////////////////

// LoadAuthInfo loads the authentication information from the file
// returning the correct type of AuthInfo concrete type.
func LoadAuthInfo(fname string) (AuthInfo, error) {
	var r io.Reader

	// Open the file and schedule it to be closed.
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the JSON file into a buffer.
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Try to create a OAuthToken.
	r = strings.NewReader(string(buf))
	oauthToken, err := NewOAuthTokenFromJSON(r)
	if err == nil {
		return &oauthToken, nil
	}

	// Try to create a PrivateToken.
	r = strings.NewReader(string(buf))
	privateToken, err := NewPrivateTokenFromJSON(r)
	if err == nil {
		return &privateToken, nil
	}

	// Try to create a BasicAuthInfo.
	r = strings.NewReader(string(buf))
	basicAuthInfo, err := NewBasicAuthInfoFromJSON(r)
	if err == nil {
		return &basicAuthInfo, nil
	}

	return nil, err
}

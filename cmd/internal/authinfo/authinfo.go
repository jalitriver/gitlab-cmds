// This file is used for reading the user's access token from a XML
// file.  The file should have one of the following formats:
//
//  <AuthInfo>
//
//    <!--
//        Select just one of the following below to specify your OAuth
//        token, private or personal token, or HTTP basic authentication.
//    -->
//
//    <!--
//        <oauth-token></private-token>
//    -->
//
//    <!--
//        <private-token></private-token>
//    -->
//
//    <!--
//        <username></username>
//        <password></password>
//    -->
//
//  </AuthInfo>

package authinfo

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/xanzy/go-gitlab"
)

////////////////////////////////////////////////////////////////////////
// Errors
////////////////////////////////////////////////////////////////////////

var (
	ErrAuthInfoInvalidXML = errors.New("invalid XML")
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
	Username string `xml:"username"`
	Password string `xml:"password"`
}

// NewBasicAuthInfo creates a new set of authentication information
// for HTTP basic authentication.
func NewBasicAuthInfo(username, password string) BasicAuthInfo {
	return BasicAuthInfo{
		Username: username,
		Password: password,
	}
}

// NewBasicAuthInfoFromXML creates a new set of authentication
// information for HTTP basic authentication from the XML accessible
// through the io.Reader.  The format of the XML is as follows:
//
//	<AuthInfo>
//	    <username></username>
//	    <password></password>
//	</AuthInfo>
func NewBasicAuthInfoFromXML(r io.Reader) (BasicAuthInfo, error) {
	result := BasicAuthInfo{}
	err := xml.NewDecoder(r).Decode(&result)
	if err != nil {
		return BasicAuthInfo{}, err
	}
	if (len(result.Username) == 0) || (len(result.Password) == 0) {
		return BasicAuthInfo{}, ErrAuthInfoInvalidXML
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
	Token string `xml:"oauth-token"`
}

// NewOAuthToken creates a new set of authentication information for
// OAuth authentication.
func NewOAuthToken(token string) OAuthToken {
	return OAuthToken{
		Token: token,
	}
}

// NewOAuthTokenFromXML creates a new set of authentication
// information for OAuth authentication from the XML accessible
// through the io.Reader.  The format of the XML is as follows:
//
//	<AuthInfo>
//	    <oauth-token></oauth-token>
//	</AuthInfo>
func NewOAuthTokenFromXML(r io.Reader) (OAuthToken, error) {
	result := OAuthToken{}
	err := xml.NewDecoder(r).Decode(&result)
	if err != nil {
		return OAuthToken{}, err
	}
	if len(result.Token) == 0 {
		return OAuthToken{}, ErrAuthInfoInvalidXML
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
	Token string `xml:"private-token"`
}

// NewPrivateToken creates a new set of authentication information for
// private token or personal token authentication.
func NewPrivateToken(token string) PrivateToken {
	return PrivateToken{
		Token: token,
	}
}

// NewPrivateTokenFromXML creates a new set of authentication
// information for private token or personal token authentication from
// the XML accessible through the io.Reader.  The format of the XML
// is as follows:
//
//	<AuthInfo>
//	    <private-token></private-token>
//	<AuthInfo>
func NewPrivateTokenFromXML(r io.Reader) (PrivateToken, error) {
	result := PrivateToken{}
	err := xml.NewDecoder(r).Decode(&result)
	if err != nil {
		return PrivateToken{}, err
	}
	if len(result.Token) == 0 {
		return PrivateToken{}, ErrAuthInfoInvalidXML
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
func Load(fname string) (AuthInfo, error) {
	var r io.Reader

	// Open the file and schedule it to be closed.
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the XML file into a buffer.
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Try to create a OAuthToken.
	r = strings.NewReader(string(buf))
	oauthToken, err := NewOAuthTokenFromXML(r)
	if err == nil {
		return &oauthToken, nil
	}

	// Try to create a PrivateToken.
	r = strings.NewReader(string(buf))
	privateToken, err := NewPrivateTokenFromXML(r)
	if err == nil {
		return &privateToken, nil
	}

	// Try to create a BasicAuthInfo.
	r = strings.NewReader(string(buf))
	basicAuthInfo, err := NewBasicAuthInfoFromXML(r)
	if err == nil {
		return &basicAuthInfo, nil
	}

	return nil, err
}

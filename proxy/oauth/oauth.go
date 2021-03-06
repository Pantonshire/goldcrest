package oauth

import (
  "bytes"
  "crypto/hmac"
  "crypto/rand"
  "crypto/sha1"
  "encoding/base64"
  "fmt"
  "github.com/martinlindhe/base36"
  "net/http"
  "path"
  "strings"
  "time"
)

const (
  oauthVersion         = "1.0"
  oauthSignatureMethod = "HMAC-SHA1"
  oauthNonceBytes      = 32
)

type Auth struct {
  Key   string `json:"key"`
  Token string `json:"token"`
}

type AuthPair struct {
  Secret Auth `json:"secret"`
  Public Auth `json:"public"`
}

type Request struct {
  Method, Protocol, Domain, Path string
  Query, Body                    Params
}

func NewRequest(method, protocol, domain, path string, query, body Params) Request {
  return Request{
    Method:   method,
    Protocol: protocol,
    Domain:   domain,
    Path:     path,
    Query:    query,
    Body:     body,
  }
}

// Creates a new http.Request containing an authentication header as described at
// https://developer.twitter.com/en/docs/authentication/oauth-1-0a/authorizing-a-request
func (or Request) MakeRequest(auth AuthPair) (*http.Request, error) {
  nonce, err := randBase36(oauthNonceBytes)
  if err != nil {
    return nil, err
  }

  baseURL := or.Protocol + "://" + path.Join(or.Domain, or.Path)

  queryParams, bodyParams := percentEncodedParams(or.Query), percentEncodedParams(or.Body)

  timestamp := fmt.Sprintf("%d", time.Now().Unix())

  oauthParams := percentEncodedParams{}
  oauthParams.set("oauth_consumer_key", auth.Public.Key)
  oauthParams.set("oauth_token", auth.Public.Token)
  oauthParams.set("oauth_signature_method", oauthSignatureMethod)
  oauthParams.set("oauth_version", oauthVersion)
  oauthParams.set("oauth_timestamp", timestamp)
  oauthParams.set("oauth_nonce", nonce)

  signature := signOAuth(auth.Secret, or.Method, baseURL, oauthParams, queryParams, bodyParams)
  oauthParams.set("oauth_signature", signature)

  authorization := "OAuth " + oauthParams.encode(", ", true)

  fullURL := baseURL + "?" + queryParams.encode("&", false)
  bodyStr := bodyParams.encode("&", false)

  req, err := http.NewRequest(or.Method, fullURL, bytes.NewBufferString(bodyStr))
  if err != nil {
    return nil, err
  }

  if bodyStr != "" {
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  }
  req.Header.Set("Authorization", authorization)

  return req, nil
}

// Creates an OAuth signature using the method described at
// https://developer.twitter.com/en/docs/authentication/oauth-1-0a/creating-a-signature
func signOAuth(secret Auth, method, baseURL string, oauthParams, queryParams, bodyParams percentEncodedParams) string {
  allParams := percentEncodedParams{}
  for key, value := range oauthParams {
    allParams.set(key, value)
  }
  for key, value := range queryParams {
    allParams.set(key, value)
  }
  for key, value := range bodyParams {
    allParams.set(key, value)
  }
  paramStr := allParams.encode("&", false)
  sigBase := strings.ToUpper(method) + "&" + percentEncode(baseURL) + "&" + percentEncode(paramStr)
  sigKey := percentEncode(secret.Key) + "&" + percentEncode(secret.Token)
  hash := hmac.New(sha1.New, []byte(sigKey))
  hash.Write([]byte(sigBase))
  return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func randBase36(bytes int) (string, error) {
  randBytes := make([]byte, bytes)
  if _, err := rand.Read(randBytes); err != nil {
    return "", err
  }
  return base36.EncodeBytes(randBytes), nil
}

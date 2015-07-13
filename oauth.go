package yelp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

//The oauth structure represents all data and provides all required method tha
//are necessary to sign a yelp api query.
type oauth struct {
	ConsumerKey string
	Token       string
	hashKey     []byte
}

//SetHashKey will create the hash key string following oauth 1.0 guidelines
func (yoa *oauth) SetHashKey(consumerSecret string, tokenSecret string) {
	//create hashing key
	var hashKey bytes.Buffer
	hashKey.WriteString(percentEncode(consumerSecret))
	hashKey.WriteString("&")
	hashKey.WriteString(percentEncode(tokenSecret))

	//set hash key
	yoa.hashKey = hashKey.Bytes()
}

//Sign will use the hash key and a HMAC-SHA1 algorithm to generate and sign a
//signature. This signature, together with all other important oauth search query
//elements, will be added to the SearchQuery.
func (yoa *oauth) Sign(method string, url string, elements *SearchQuery) {
	//add the OAuth elements to the Yelp query
	elements.Append("oauth_consumer_key", yoa.ConsumerKey)
	elements.Append("oauth_nonce", nonce(30))
	elements.Append("oauth_signature_method", "HMAC-SHA1")
	elements.Append("oauth_timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	elements.Append("oauth_token", yoa.Token)

	//sort all query elements
	elements.Sort()

	//create the signature
	signature := strings.Join([]string{method, percentEncode(url), percentEncode(elements.String())}, "&")

	//reset the hasher (could have been used before), then sign the signature
	//yoa.Hasher.Reset()
	hasher := hmac.New(sha1.New, yoa.hashKey)
	hasher.Write([]byte(signature))

	//add the signature to the query and percent encode it
	elements.Append("oauth_signature", percentEncode(base64.StdEncoding.EncodeToString(hasher.Sum(nil))))
}

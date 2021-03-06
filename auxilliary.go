package yelp

import (
	"bytes"
	"math/rand"
)

//The hexMap string is used by the percentEncode function to encode a character
//that is not allowed to exist as plaintext.
const hexMap string = "0123456789ABCDEF"

//shouldPercentEncode returns true if a character should be percent encoded
func shouldPercentEncode(c byte) bool {
	return !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '-' || c == '.' || c == '_' || c == '~')
}

//percentEncode will percent encode a provided string and return it.
func percentEncode(source string) string {
	var buffer bytes.Buffer

	for _, v := range source {
		val := byte(v)
		if shouldPercentEncode(val) {
			//I know this is the weirdest hack ever. But somehow Yelp does not like
			//it when it has to percent encode a comma. This should be %2C, but yelp
			//expects it to be %252C (coincedentally, %25 == '%', maybe a double
			//URL encoding error on their part?).
			if val == ',' {
				buffer.WriteString("%252C")
			} else {
				buffer.WriteByte('%')
				buffer.WriteByte(hexMap[(val>>4)&0x0F])
				buffer.WriteByte(hexMap[val&0x0F])
			}
		} else {
			buffer.WriteByte(val)
		}
	}

	return buffer.String()
}

//The nonceCharacter string contains all allowed letters that can be used when
//generating a new nonce string
const nonceCharacters string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const nonceLength int = len(nonceCharacters)

//nonce will return a nonce-string with a specified length in characters
func nonce(length int) string {
	var result bytes.Buffer

	for i := 0; i < length; i++ {
		result.WriteByte(nonceCharacters[rand.Int()%nonceLength])
	}

	return result.String()
}

//validLatitudeLongitude returns a boolean value that is true when the provided
//latitude and longitude are valid. It is false when one of these values is
//invalid
func validLatitudeLongitude(latitude, longitude float64) bool {
	return !(latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180)
}

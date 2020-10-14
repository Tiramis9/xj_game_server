/*
 * @Author: yhlyl
 * @Date: 2019-10-22 11:33:37
 * @LastEditTime: 2019-11-04 21:28:07
 * @LastEditors: yhlyl
 * @Description:
 * @FilePath: /gin_micro/util/jwt/jwt.go
 * @Github: https://github.com/android-coco/gin_micro
 */
package jwt

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// EasyToken is an Struct to encapsulate username and expires as parameter
type EasyToken struct {
	// Username is the name of the user
	Username string
	// Expires is a timestamp with expiration date
	Expires int64
}

// https://gist.github.com/cryptix/45c33ecf0ae54828e63b
// location of the files used for signing and verification
const (
	privKeyPath = "/../conf/rsa_private_key.pem" // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "/../conf/rsa_public_key.pem"  // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

var (
	verifyKey    *rsa.PublicKey
	mySigningKey *rsa.PrivateKey
)

func init() {
	pathStr, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	verifyBytes, err := ioutil.ReadFile(pathStr + pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}

	signBytes, err := ioutil.ReadFile(pathStr + privKeyPath)

	if err != nil {
		log.Fatal(err)
	}

	mySigningKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// GetToken is a function that exposes the method to get a simple token for jwt
func (e EasyToken) GetToken() (string, error) {

	// Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: e.Expires, //time.Unix(c.ExpiresAt, 0)
		Issuer:    e.Username,
		NotBefore: int64(time.Now().Unix()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		log.Fatal(err)
	}

	return tokenString, err
}

// ValidateToken get token strings and return if is valid or not
func (e EasyToken) ValidateToken(tokenString string) (bool, string, error) {
	// Token from another example.  This token is expired
	//var tokenString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJleHAiOjE1MDAwLCJpc3MiOiJ0ZXN0In0.HE7fK0xOQwFEr4WDgRWj4teRPZ6i3GLwD5YCm6Pwu_c"
	if tokenString == "" {
		return false, "", errors.New("token is empty")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	if token == nil {
		log.Println(err)
		return false, "", errors.New("not work")
	}

	if token.Valid {
		//"You look nice today"
		claims, _ := token.Claims.(jwt.MapClaims)
		//var user string = claims["username"].(string)
		iss := claims["iss"].(string)
		return true, iss, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return false, "", errors.New("that's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			return false, "", errors.New("timing is everything")
		} else {
			//"Couldn't handle this token:"
			return false, "", err
		}
	} else {
		//"Couldn't handle this token:"
		return false, "", err
	}
}

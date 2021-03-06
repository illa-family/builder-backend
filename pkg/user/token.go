// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package user

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type AuthClaims struct {
	User   int    `json:"user"`
	Random string `json:"rnd"`
	jwt.RegisteredClaims
}

func CreateAccessToken(id int) (string, error) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vCode := fmt.Sprintf("%06v", rnd.Int31n(10000))

	claims := &AuthClaims{
		User:   id,
		Random: vCode,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "ILLA",
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Hour * 24 * 7),
			},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(os.Getenv("ILLA_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func ValidateAccessToken(accessToken string) (bool, error) {
	_, err := ExtractUserIDFromToken(accessToken)
	if err != nil {
		return false, err
	}
	return true, nil
}

func ExtractUserIDFromToken(accessToken string) (int, error) {
	authClaims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, authClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ILLA_SECRET_KEY")), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !(ok && token.Valid) {
		return 0, err
	}

	userID := claims.User

	return userID, nil
}

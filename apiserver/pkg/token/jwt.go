package token

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

var (
	jwtKey []byte
)

func NewJwtKey(key string) {
	jwtKey = []byte(key)
}

// Encode generate tokens used for auth
func Encode(uid int64) string {
	date := time.Now().Unix()
	if jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": date,           //处理时间,在此时间前不可用
		"iat": date,           //签发时间
		"exp": date + 1209600, //到期时间
		//"exp": date + 20, //到期时间
		"iss": "MuxWaf", //签发人,iss
		"uid": uid,
	}).SignedString(jwtKey); err == nil {
		return jwtToken
	}
	return ""
}

// Decode parsing token
//func Decode(tokenString string) (string, bool) {
//	if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//		}
//		return jwtKey, nil
//	}); err == nil {
//		if value, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//			if v, ok := value["uid"].(string); ok {
//				return v, true
//			}
//		}
//	}
//	return "", false
//}

/// Decode parsing token
func Decode(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return claims, nil
		} else {
			// TODO 日志
			return nil, err
		}
	}
	return nil, err

}

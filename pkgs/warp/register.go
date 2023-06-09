package warp

import (
	"bytes"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gogf/gf/encoding/gjson"
	"golang.org/x/crypto/curve25519"
)

/*
`{"key":"` + publicKey + `","install_id":"` + installID + `","fcm_token":"` + installID + `:APA91b` + fcmtoken + `","tos":"` + time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + `","model":"Android","serial_number":"` + installID + `","locale":"zh_CN"}`
*/
const (
	WRPayLoad string = `{"key":"%s","install_id":"%s","fcm_token":"%s:APA91b%s","tos":"%s","model":"PC","serial_number":"%s","locale":"zh_CN"}`
)

func GenRandomStr(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenKey() (privateKey, publicKey string, err error) {
	b := make([]byte, 32)

	if _, err := crand.Read(b); err != nil {
		return "", "", fmt.Errorf("cannot read random bytes: %v", err)
	}
	b[0] &= 248
	b[31] &= 127
	b[31] |= 64
	var pub, priv [32]byte
	copy(priv[:], b)
	curve25519.ScalarBaseMult(&pub, &priv)
	return base64.StdEncoding.EncodeToString(priv[:]), base64.StdEncoding.EncodeToString(pub[:]), nil
}

func ParseReceived(clientID string) (r []int) {
	decoded, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		fmt.Println(err)
		return
	}
	hexString := hex.EncodeToString(decoded)

	reserved := []int{}
	for i := 0; i < len(hexString); i += 2 {
		hexByte := hexString[i : i+2]
		decValue, _ := strconv.ParseInt(hexByte, 16, 64)
		reserved = append(reserved, int(decValue))
	}
	r = reserved
	return
}

type WarpRegister struct {
	Url       string
	Headers   map[string]string
	installID string
	fcmToken  string
}

func NewWarpRegister() (w *WarpRegister) {
	w = &WarpRegister{
		Url: "https://api.cloudflareclient.com/v0a2158/reg",
		Headers: map[string]string{
			"CF-Client-Version": "a-6.10-2158",
			"User-Agent":        "okhttp/3.12.1",
			"Content-Type":      "application/json; charset=UTF-8",
		},
	}
	w.installID = GenRandomStr(22)
	w.fcmToken = GenRandomStr(134)
	return
}

func (that *WarpRegister) Register() error {
	privateKey, publicKey, err := GenKey()
	if err != nil {
		return err
	}
	payload := fmt.Sprintf(
		WRPayLoad,
		publicKey,
		that.installID,
		that.installID,
		that.fcmToken,
		time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		that.installID,
	)
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12},
	}}
	req, err := http.NewRequest(http.MethodPost, that.Url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	for k, v := range that.Headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("PrivateKey: ", privateKey)
	fmt.Println(string(content))
	j := gjson.New(content)
	clientID := j.GetString("config.client_id")
	r := ParseReceived(clientID)
	fmt.Println(r)
	return nil
}

func TestWarp() {
	wp := NewWarpRegister()
	wp.Register()
}

/*
{
	"id": "f50d73af-39e3-4aca-820f-c750a52a82ca",
	"type": "a",
	"model": "Unknown Android SDK built for x86",
	"name": "",
	"key": "JUrMi7sQM2O42sb0i8TW+XAMXARAIikWCNIyHV4Suws=",
	"account": {
		"id": "d83e2939-6df2-4d9f-9e87-7d68135f7ed5",
		"account_type": "free",
		"created": "2023-04-24T16:04:43.609189124Z",
		"updated": "2023-04-24T16:04:43.609189124Z",
		"premium_data": 0,
		"quota": 0,
		"usage": 0,
		"warp_plus": true,
		"referral_count": 0,
		"referral_renewal_countdown": 0,
		"role": "child",
		"license": "35y80YQc-172UxqH8-928WOLS6"
	},
	"config": {
		"client_id": "VYX7",
		"peers": [{
			"public_key": "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
			"endpoint": {
				"v4": "162.159.192.6:0",
				"v6": "[2606:4700:d0::a29f:c006]:0",
				"host": "engage.cloudflareclient.com:2408"
			}
		}],
		"interface": {
			"addresses": {
				"v4": "172.16.0.2",
				"v6": "2606:4700:110:8bd5:acd4:56b1:e443:ace7"
			}
		},
		"services": {
			"http_proxy": "172.16.0.1:2480"
		}
	},
	"token": "159ce275-d53a-4fb0-aaa9-159232366f54",
	"warp_enabled": false,
	"waitlist_enabled": false,
	"created": "2023-04-24T16:04:43.231171579Z",
	"updated": "2023-04-24T16:04:43.231171579Z",
	"tos": "2023-04-24T16:04:42.013Z",
	"place": 0,
	"locale": "zh-CN",
	"enabled": true,
	"install_id": "diF0aHXdTY2paujd3NbtLx",
	"fcm_token": "diF0aHXdTY2paujd3NbtLx:APA91bEv8VREoAe7WK951PiK0-h7ZoZz9aTmaz3O_z8vS5zPrOC29OdHabNNLLIHE7uRoHhqcXy3lePyCZq5ysahblEOC8NpQWqMbPzCjQRoNZs6FVcrRNbX70hZQSnIv4H4cHD5Cvtn",
	"serial_number": "diF0aHXdTY2paujd3NbtLx"
}
*/

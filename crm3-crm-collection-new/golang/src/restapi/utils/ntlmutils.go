package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf16"

	"github.com/julienschmidt/httprouter"
)

func decodeUTF16(b []byte, order binary.ByteOrder) (string, error) {
	u16s := []uint16{}

	for i, j := 0, len(b); i < j; i += 2 {
		u16s = append(u16s, order.Uint16(b[i:]))
	}

	runes := utf16.Decode(u16s)
	return string(runes), nil
}

func NTLMProcess(res http.ResponseWriter, req *http.Request, param httprouter.Params) (string, string, string, error) {
	auth := req.Header.Get("Authorization")
	log.Println("auth", auth)
	if !strings.HasPrefix(auth, "NTLM") {
		res.Header().Add("WWW-Authenticate", "NTLM")
		res.WriteHeader(401)
		return "", "", "", errors.New("Error NTLMProcess")
	} else {
		arr := strings.Split(auth, "NTLM ")
		b, _ := base64.StdEncoding.DecodeString(arr[1])
		log.Println("NTLM type", int(b[8]))
		if int(b[8]) == 1 {
			host := (fmt.Sprintf("%s", b[int(b[28]):int(b[28])+int(b[16])]))
			domain := (fmt.Sprintf("%s", b[int(b[20]):int(b[20])+int(b[18])]))
			log.Println("host:", host)
			log.Println("domain:", domain)
			str := "4e544c4d53535000020000000000000028000000018200005372764e6f6e63650000000000000000"
			decoded, _ := hex.DecodeString(str)
			fmt.Printf("%s\n", decoded)
			log.Println(base64.StdEncoding.EncodeToString(decoded))
			res.Header().Add("WWW-Authenticate", "NTLM "+base64.StdEncoding.EncodeToString(decoded))
			res.WriteHeader(401)
			return "", "", "", errors.New("Access Denied")
		} else if int(b[8]) == 3 {

			domain, _ := decodeUTF16(b[int(b[32]):int(b[32])+int(b[28])], binary.LittleEndian)
			user, _ := decodeUTF16(b[int(b[40]):int(b[40])+int(b[36])], binary.LittleEndian)
			host, _ := decodeUTF16(b[int(b[48]):int(b[48])+int(b[44])], binary.LittleEndian)

			log.Println("host 3:", host)
			log.Println("domain 3:", domain)
			log.Println("user 3:", user)

			SetSessionData(req, "NTLM_DOMAIN", domain)
			SetSessionData(req, "NTLM_USER", user)
			SetSessionData(req, "NTLM_HOST", host)
			return domain, user, host, nil

			/*res.Header().Add("Content-Type", "text/html; charset=utf-8")

			res.Write([]byte(domain + "<br />"))
			res.Write([]byte(user + "<br />"))
			res.Write([]byte(host + "<br />"))
			*/

		}
	}
	return "", "", "", nil
}

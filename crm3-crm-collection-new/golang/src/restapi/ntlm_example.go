package restapi

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf16"

	"github.com/julienschmidt/httprouter"
)

//https://www.innovation.ch/personal/ronald/ntlm.html#hashes

type type1Message struct {
	protocol [8]byte
	type_    byte // 0x01
	zero1    [3]byte
	flags    [2]byte // 0xb203
	zero2    [2]byte

	dom_len1 [2]byte // domain string length
	dom_len2 [2]byte // domain string length
	dom_off  [2]byte // domain string offset
	zero3    [2]byte

	host_len1 [2]byte // host string length
	host_len2 [2]byte // host string length
	host_off  [2]byte // host string offset (always 0x20)
	zero4     [2]byte

	host []byte // host string (ASCII)
	dom  []byte // domain string (ASCII)
}

func decodeUTF16(b []byte, order binary.ByteOrder) (string, error) {
	u16s := []uint16{}

	for i, j := 0, len(b); i < j; i += 2 {
		u16s = append(u16s, order.Uint16(b[i:]))
	}

	runes := utf16.Decode(u16s)
	return string(runes), nil
}

func NTLMExample(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "NTLM") {
		res.Header().Add("WWW-Authenticate", "NTLM")
		res.WriteHeader(401)
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
		} else if int(b[8]) == 3 {
			domain, _ := decodeUTF16(b[int(b[32]):int(b[32])+int(b[28])], binary.LittleEndian)
			user, _ := decodeUTF16(b[int(b[40]):int(b[40])+int(b[36])], binary.LittleEndian)
			host, _ := decodeUTF16(b[int(b[48]):int(b[48])+int(b[44])], binary.LittleEndian)

			res.Header().Add("Content-Type", "text/html; charset=utf-8")

			res.Write([]byte(domain + "<br />"))
			res.Write([]byte(user + "<br />"))
			res.Write([]byte(host + "<br />"))

		}
	}

	return

}

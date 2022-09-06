package http

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/util"

	lua "github.com/Shopify/go-lua"
)

func luaPost(l *lua.State) int {

	timeout := time.Duration(30000 * time.Second)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	url := lua.CheckString(l, 1)
	body := lua.CheckString(l, 2)
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))

	headersTable, err := util.PullInterfaceTable(l, 3)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}

	for k, v := range headersTable {
		req.Header.Set(k, v.(string))
		//log.Println("key=>", k, "value=>", v)
	}

	resp, err := client.Do(req)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString(string(response))
	l.PushNil()
	l.PushInteger(0)
	return 3
}

var lib = []lua.RegistryFunction{
	{"post", luaPost},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/http", libOpen, false)
	l.Pop(1)
}

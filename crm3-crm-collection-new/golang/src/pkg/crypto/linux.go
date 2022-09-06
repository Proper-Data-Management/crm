// +build linux

package crypto

import (
	"git.dar.kz/crediton-3/crm-mfo/src/lib/gokalkan"
	lua "github.com/Shopify/go-lua"
)

func luaSignXml(l *lua.State) int {
	xmlInData := lua.CheckString(l, 1)
	signNodeXpath := lua.CheckString(l, 2)
	data, err := gokalkan.SignXml([]byte(xmlInData), signNodeXpath)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString(string(data))
	l.PushNil()
	l.PushInteger(0)
	return 3
}

func luaVerifyXml(l *lua.State) int {
	xmlInData := lua.CheckString(l, 1)
	err := gokalkan.VerifyXml([]byte(xmlInData))
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 2
	}
	l.PushNil()
	l.PushInteger(0)
	return 2
}
func luaLoadKey(l *lua.State) int {
	container := lua.CheckString(l, 1)
	password := lua.CheckString(l, 2)
	err := gokalkan.LoadKey(container, password)
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 2
	}
	l.PushNil()
	l.PushInteger(0)
	return 2
}
func luaTSASetUrl(l *lua.State) int {
	url := lua.CheckString(l, 1)
	err := gokalkan.TSASetUrl(url)
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 2
	}
	l.PushNil()
	l.PushInteger(0)
	return 2
}

func luaGetCertFromXml(l *lua.State) int {
	xmlInData := lua.CheckString(l, 1)
	data, err := gokalkan.GetCertFromXml([]byte(xmlInData))
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString(string(data))
	l.PushNil()
	l.PushInteger(0)
	return 3
}

func luaLoadCertificateFromFile(l *lua.State) int {
	filename := lua.CheckString(l, 1)
	certType := lua.CheckInteger(l, 2)
	err := gokalkan.LoadCertificateFromFile(filename, certType)
	if err != nil {
		l.PushString(err.Error())
		l.PushInteger(1)
		return 2
	}
	l.PushNil()
	l.PushInteger(0)
	return 2
}

func luaCertificateGetInfo(l *lua.State) int {
	inCert := lua.CheckString(l, 1)
	propId := lua.CheckInteger(l, 2)
	data, err := gokalkan.CertificateGetInfo([]byte(inCert), propId)
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		l.PushInteger(1)
		return 3
	}
	l.PushString(string(data))
	l.PushNil()
	l.PushInteger(0)
	return 3
}

var lib = []lua.RegistryFunction{
	{"LoadKey", luaLoadKey},
	{"SignXml", luaSignXml},
	{"VerifyXml", luaVerifyXml},
	{"TSASetUrl", luaTSASetUrl},
	{"GetCertFromXml", luaGetCertFromXml},
	{"LoadCertificateFromFile", luaLoadCertificateFromFile},
	{"CertificateGetInfo", luaCertificateGetInfo},
}

func Open(l *lua.State) {
	libOpen := func(l *lua.State) int {
		lua.NewLibrary(l, lib)
		return 1
	}
	lua.Require(l, "pkg/crypto", libOpen, false)
	l.Pop(1)
}

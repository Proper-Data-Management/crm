package pkg

import (
	"git.dar.kz/crediton-3/crm-mfo/src/lib/lua/dbrequire"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/ast"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/crypto"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/http"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/json"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/jsonpath"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/lock"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/mxj"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/path/filepath"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/strings"
	"git.dar.kz/crediton-3/crm-mfo/src/pkg/xml"
	"github.com/Shopify/go-lua"
)

func Open(l *lua.State) {
	dbrequire.Open(l)
	crypto.Open(l)
	xml.Open(l)
	jsonpath.Open(l)
	mxj.Open(l)
	http.Open(l)
	ast.Open(l)
	lock.Open(l)
	strings.Open(l)
	json.Open(l)
	filepath.Open(l)

}

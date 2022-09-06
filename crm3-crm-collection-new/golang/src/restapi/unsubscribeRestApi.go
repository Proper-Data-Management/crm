package restapi

import (
	"fmt"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/julienschmidt/httprouter"
)

func EmailUnSubscribe(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")
	email, email_from := "", ""
	o.Raw(utils.DbBindReplace("select email,email_from from email_logs where sys$uuid=?"), param.ByName("uuid")).QueryRow(&email, &email_from)
	o.Raw(utils.DbBindReplace("insert into email_unsubscribes (email,email_from) values (?,?)"), email, email_from).Exec()

	fmt.Fprint(res, `
<html>
<head>
<title>Вы успешно отписались от рассылки</title>
<meta charset="utf-8"/>
</head>
<body>
<code>`+email+`,Вы успешно отписались от рассылки от адреса `+email_from+`</code>
</body>
</html>
	`)
}

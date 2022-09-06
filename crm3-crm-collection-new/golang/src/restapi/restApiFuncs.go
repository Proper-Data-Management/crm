package restapi
import (
	"net/http"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
)

func IsMobile (req *http.Request) bool {
	return utils.System(req) == "android" || utils.System(req) == "ios"
}

package restapi
import (
	"net/http"
	"encoding/json"
	"fmt"
)

type TOKResponse struct {
	OKText string `json:"ok_text"`
	Error string `json:"error"`
}
func OkResponse(res http.ResponseWriter,text string ) {

	okRes := TOKResponse{Error:"0",OKText:text}
	jsonData, _ := json.Marshal(okRes)
	fmt.Fprint(res, string(jsonData))
}

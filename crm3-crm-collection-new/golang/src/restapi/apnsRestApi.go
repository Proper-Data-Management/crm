package restapi
import (
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"net/http"
	"fmt"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/apns"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"


)

func ApnsRestApi(res http.ResponseWriter, req *http.Request, _ httprouter.Params){


	type apnSRequest struct {
		DeviceToken string `json:"deviceToken"`
		Alert string `json:"alert"`
		Badge int `json:"badge"`
		Sound string `json:"sound"`
	}
	type apnSResponse struct {
		Success bool `json:"success"`
		Error string `json:"error"`

	}

	decoder := json.NewDecoder(req.Body)
	var request apnSRequest
	err := decoder.Decode(&request)
	if err != nil {
		RestCheckPanic(err,res)
		return
	}

	payload := apns.NewPayload()
	payload.Alert = request.Alert
	payload.Badge = request.Badge
	payload.Sound = request.Sound

	pn := apns.NewPushNotification()
	pn.DeviceToken = request.DeviceToken
	pn.AddPayload(payload)

	client := apns.NewClient(utils.GetParamValue("apns_sandbox_url"), utils.GetParamValue("ntmy_pem_path"), utils.GetParamValue("ntmy_pem_path"))
	resp := client.Send(pn)

	alert, _ := pn.PayloadString()
	fmt.Println("  Alert:", alert)
	fmt.Println("Success:", resp.Success)
	fmt.Println("  Error:", resp.Error)

	var result apnSResponse
	result.Success = resp.Success
	if resp.Error!=nil {
		result.Error = resp.Error.Error()
	}

	jsonData, _ := json.Marshal(result)
	fmt.Fprint(res, string(jsonData))
}

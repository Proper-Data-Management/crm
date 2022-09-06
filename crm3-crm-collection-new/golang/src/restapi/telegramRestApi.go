package restapi

import "net/http"
import (
	"encoding/json"
	"io/ioutil"
	"log"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/julienschmidt/httprouter"
)

func TelegramWebHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var input []luautils.NameValue
	r.ParseForm()
	bp_id := int64(0)
	token := r.Form.Get("token")
	//log.Println(token)

	o := orm.NewOrm()
	o.Using("default")

	user_id := int64(0)
	err := o.Raw(utils.DbBindReplace("select bp_id,user_id from tg_bots where token=?"), token).QueryRow(&bp_id, &user_id)
	if err != nil {
		RestCheckPanic(err, w)
		return
	}

	//log.Println("Norm")

	updatesChan := make(chan tgbotapi.Update, 100)
	bytes, _ := ioutil.ReadAll(r.Body)

	//s := string(bytes)
	//log.Println(s)

	var update tgbotapi.Update
	json.Unmarshal(bytes, &update)
	//log.Println(update.Message.Text)
	updatesChan <- update

	if update.Message == nil {
		log.Println("update message nil", update)
		return
	}

	update.Message.Entities = nil
	upd, err := json.Marshal(update)
	if err != nil {
		log.Println("TelegramWebHook 1 ", err)
		return
	}
	//log.Println("upd=" , upd)

	input = append(input, luautils.NameValue{Name: "input", Value: string(upd)})
	input = append(input, luautils.NameValue{Name: "token", Value: token})
	//log.Println(bp_id)

	instanceContext := luautils.InstanceContext{O: o}

	_, _, _, _, err = instanceContext.CreateInstance(r, bp_id, user_id, input, 0)

	//log.Println(output)
	//log.Println(instanceId)
	if err != nil {
		log.Println("TelegramWebHook 2 ", err)
		return
	}

	u := tgbotapi.NewUpdate(0)

	u.Timeout = 60

	/*bot, err := tgbotapi.NewBotAPI(token)
	//msg := tgbotapi.NewDocumentShare(update.Message.Chat.ID, "BQADAgADQAADw3xXB8_ac4mTPXDJAg")
	kbd := tgbotapi.NewKeyboardButtonLocation("Где ты?");
	var kbds []tgbotapi.KeyboardButton
	kbds = append(kbds,kbd)
	var kbdsKB []tgbotapi.InlineKeyboardButton
	//kbdsKB = append(kbdsKB,kbds)

	tgbotapi.NewInlineKeyboardMarkup()

	var r = tgbotapi.NewEditMessageReplyMarkup(update.Message.Chat.ID,update.Message.MessageID,kbdsKB )
	bot.Send(r)
	if err!=nil{
		log.Println(err)
	}

	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60

	bot, err := tgbotapi.NewBotAPI(token)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Мир!")
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
	*/
}

func TelegramResetWebHooks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	o := orm.NewOrm()
	o.Using("default")

	r.ParseForm()
	token := ""
	err := o.Raw(utils.DbBindReplace("select token from tg_bots where token=?"), r.Form.Get("token")).QueryRow(&token)

	if err != nil {
		RestCheckPanic(err, w)
		return
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Println(err.Error() + " " + token)
		RestCheckPanic(err, w)
		return
	}
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(""))
	if err != nil {
		log.Println("test3")
		RestCheckPanic(err, w)
		return
	}
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(utils.GetParamValue("telegram_webhook_uri") + "?token=" + bot.Token))
	if err != nil {
		log.Println("test4")
		RestCheckPanic(err, w)
		return
	}
}

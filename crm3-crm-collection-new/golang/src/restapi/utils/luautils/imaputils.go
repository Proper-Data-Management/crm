package luautils

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"

	lua "github.com/Shopify/go-lua"
	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func ImapProcessInBoxToMessage(L *lua.State, callBack string,

	pullCount float64,
	lastUid float64,

	login string, password string, hostPort string, user_id float64) (float64, error) {
	log.Println("Connecting to server...")

	type timapFileInputMessage struct {
		FileName string `json:"Filename"`
		Body     []byte `json:"Body"`
	}

	type imapFileInputMessage timapFileInputMessage
	type imapInputMessage struct {
		UID     uint32                 `json:"uid"`
		Subject string                 `json:"subject"`
		To      []*mail.Address        `json:"to"`
		From    []*mail.Address        `json:"from"`
		Body    string                 `json:"body"`
		Date    string                 `json:"date"`
		Files   []imapFileInputMessage `json:"files"`
		UserId  int64                  `json:"userId"`
	}

	maxUid := uint32(0)
	// Connect to server
	c, err := client.DialTLS(hostPort, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		//log.Fatal(err)
		return 0, err
	}
	if os.Getenv("CRM_DEBUG_MAIL") == "1" {
		log.Println("Connected")
	}

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(login, password); err != nil {
		//log.Fatal(err)
		return 0, err
	}
	if os.Getenv("CRM_DEBUG_MAIL") == "1" {
		log.Println("Logged in")
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		//log.Fatal(err)
		return 0, err
	}

	// Get the last message
	if mbox.Messages == 0 {
		//log.Fatal("No message in mailbox")
		return 0, errors.New("No message in mailbox")
	}

	/*from := uint32(1)
	to := mbox.Messages

	if mbox.Messages > uint32(lastCount-1) {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - uint32(lastCount-1)
	}
	if os.Getenv("CRM_DEBUG_MAIL") == "1" {
		log.Println("from", from)
		log.Println("to", to)
	}*/

	seqSet := new(imap.SeqSet)
	//seqSet.AddNum(mbox.Messages)
	seqSet.AddRange(uint32(lastUid), uint32(lastUid+pullCount))

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, uint32(pullCount)+5)
	go func() {
		if err := c.UidFetch(seqSet, items, messages); err != nil {
			//log.Fatal(err)
			log.Println("error on fetch", err)
			return
		}
	}()

	//msg := <-messages

	processed := 0
	for msg := range messages {
		if os.Getenv("CRM_DEBUG_MAIL") == "1" {
			log.Println("Current UID", msg.Uid)
		}

		if maxUid < msg.Uid {
			maxUid = msg.Uid
		}

		if msg == nil {
			return 0, errors.New("Server didn't returned message")

		}
		body := ""
		subject := ""
		dateStr := ""
		uid := msg.Uid
		var files []imapFileInputMessage
		//var toArr []string
		var to []*mail.Address
		var from []*mail.Address
		var header mail.Header

		//log.Println("Uid", msg.Envelope.MessageId)
		r := msg.GetBody(section)

		if r == nil {
			//log.Fatal("Server didn't returned message body")
			return 0, errors.New("Server didn't returned message body")
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			//log.Fatal(err)
			log.Println("error on CreateReader ", err)
			return 0, err
		}

		// Print some info about the message
		header = mr.Header

		if date, err := header.Date(); err == nil {

			dateStr = date.Format("2006-01-02 15:04:05")

			/*if fromDate != "" {
				//utcLocation, _ := time.LoadLocation("UTC")
				//t, err := time.Parse("2006-01-02 15:04:05", fromDate)

				if dateStr < fromDate {

					//utc, _ := time.LoadLocation("Asia/Almaty")
					if os.Getenv("CRM_DEBUG_MAIL") == "1" {
						log.Println("Skipped date:", dateStr, fromDate)
					}
					//log.Println("Skipped date:", date.Local(), t)
					continue
				} else if err != nil {
					log.Println("error on parse date fromDate", fromDate)
				} else {
					//log.Println("ok date:", dateStr, fromDate)
					//log.Println("ok date:", date, t.UTC())
				}

			}*/

		}

		processed++

		if from, err = header.AddressList("From"); err == nil {
			if os.Getenv("CRM_DEBUG_MAIL") == "1" {
				log.Println("From:", from)
			}
		}
		if to, err = header.AddressList("To"); err == nil {
			if os.Getenv("CRM_DEBUG_MAIL") == "1" {
				log.Println("To:", to)
			}

			/*for tk, tv := range to {
				log.Println("to:::", tk, tv)
				toArr = append(toArr, tv.Address)
			}*/

		}
		if subject, err = header.Subject(); err == nil {
			if os.Getenv("CRM_DEBUG_MAIL") == "1" {
				log.Println("Subject:", subject)
			}
		}

		// Process each message's part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				//log.Fatal(err)
				log.Println("ImapProcessInBoxToMessage error on NextPart ", err)
				//return 0, err
				continue
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				//ioutil.ReadAll(p.Body)

				dataBody, _ := ioutil.ReadAll(p.Body)
				body = string(dataBody)
				//log.Println("Got body", body)
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				b, _ := ioutil.ReadAll(p.Body)
				files = append(files, imapFileInputMessage{FileName: filename, Body: b})
				if os.Getenv("CRM_DEBUG_MAIL") == "1" {
					log.Println("Got attachment:", filename, len(string(b)))
				}
			}
		}

		//log.Println("FILEEEES GO", files)
		//body = ""

		luaMessage := &imapInputMessage{UID: uid, Subject: subject, Date: dateStr, To: to, From: from,

			Body:   body,
			Files:  files,
			UserId: int64(user_id),
		}
		b, err := json.Marshal(luaMessage)

		var face interface{}
		err = json.Unmarshal(b, &face)

		if os.Getenv("CRM_DEBUG_MAIL") == "1" {
			//log.Println("face", string(b))
		}

		Open(L)
		DeepPush(L, face)
		L.SetGlobal("input")
		err = lua.DoString(L, callBack)
		if err != nil {
			log.Println("error on DoString ", err)
			return 0, err
		}

	}
	if os.Getenv("CRM_DEBUG_MAIL") == "1" {
		log.Println("Done ", processed)
		log.Println("MaxUID ", maxUid)
	}
	return float64(maxUid), nil
}

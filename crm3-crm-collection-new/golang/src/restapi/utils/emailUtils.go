package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
)

type TypeEmail struct {
	FromText     string
	FromMail     string
	ToText       string
	ToMail       string
	Subject      string
	Body         string
	SMTPLogin    string
	SMTPPassword string
	SMTPServer   string
	AsHTML       bool
	IsTLS        bool
}

func SendEmailTLS(email TypeEmail) error {

	from := mail.Address{email.FromText, email.FromMail}
	to := mail.Address{email.ToText, email.ToMail}
	subj := email.Subject
	body := email.Body

	// Setup headers

	o := orm.NewOrm()
	uuid := ""
	Lid, err := o.Raw(DbBindReplace("insert into email_logs (title,email) values (?,?)"), email.Subject, email.ToMail).Exec()
	LLid, err := Lid.LastInsertId()
	o.Raw(DbBindReplace("select sys$uuid from email_logs where id=?"), LLid).QueryRow(&uuid)

	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["List-Unsubscribe"] = GetParamValue("cloud_url") + "restapi/email/unsubscribe/" + uuid
	headers["To"] = to.String()
	headers["Subject"] = subj
	if email.AsHTML {
		headers["Content-Type"] = "text/html; charset =\"UTF-8\";"
	}

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	serverName := email.SMTPServer

	host, _, _ := net.SplitHostPort(serverName)

	auth := smtp.PlainAuth("", email.SMTPLogin, email.SMTPPassword, host)

	// TLS config
	/*tlsconfig := &tls.Config {

		InsecureSkipVerify: true,
		ServerName: host,

	}*/

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", serverName, nil)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	c = nil
	return nil

}

func SendEmailNoTLS(email TypeEmail) error {

	from := mail.Address{email.FromText, email.FromMail}
	to := mail.Address{email.ToText, email.ToMail}
	subj := email.Subject
	body := email.Body

	// Setup headers

	o := orm.NewOrm()
	uuid := ""
	Lid, err := o.Raw(DbBindReplace("insert into email_logs (title,email) values (?,?)"), email.Subject, email.ToMail).Exec()
	LLid, err := Lid.LastInsertId()
	o.Raw(DbBindReplace("select sys$uuid from email_logs where id=?"), LLid).QueryRow(&uuid)

	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["List-Unsubscribe"] = GetParamValue("cloud_url") + "restapi/email/unsubscribe/" + uuid
	headers["To"] = to.String()
	headers["Subject"] = subj
	if email.AsHTML {
		headers["Content-Type"] = "text/html; charset =\"UTF-8\";"
	}

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	serverName := email.SMTPServer

	host, _, _ := net.SplitHostPort(serverName)

	auth := smtp.PlainAuth("", email.SMTPLogin, email.SMTPPassword, host)

	log.Println("serverName=" + serverName)
	c, err := smtp.Dial(serverName)

	if err != nil {
		return err
	}

	// Auth

	if email.SMTPPassword != "" {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	c = nil
	return nil

}

func SendEmail(email TypeEmail) error {
	if email.IsTLS {
		return SendEmailTLS(email)
	} else {
		return SendEmailNoTLS(email)
	}
}

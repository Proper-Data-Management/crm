package luautils

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"

	//amqp "git.dar.kz/crediton-3/crm-mfo/src/restapi/mq/amqp"
	amqp "github.com/streadway/amqp"
	lua "github.com/Shopify/go-lua"
	"sync/atomic"
)

//beta
var ops int32

var amqpUri = flag.String("r", "amqp://guest:guest@127.0.0.1/", "RabbitMQ URI")

func TestSend(connStr string, queue string, exchange string, routingKey string, body string, durable bool) error {

	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Println("%s: %s", "Failed to connect to RabbitMQ", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		log.Println("%s: %s", "Failed to open a channel", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue,   // name
		durable, // durable
		false,   // delete when unused
		false,   // exclusive
		true,    // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println("%s: %s", "Failed to declare a queue", err)
		return err
	}

	pub := amqp.Publishing{}
	err = json.Unmarshal([]byte(body), &pub)
	pub.Timestamp = time.Now()
	pub.DeliveryMode = amqp.Persistent
	if err != nil {
		log.Println("Error parse", err, body)
		return err
	}

	err = ch.Publish(
		exchange, // exchange
		q.Name,   // routing key
		false,    // mandatory
		false,    // immediate
		pub,
	)
	if err != nil {
		log.Println("%s: %s", "Failed to publish a message", err)
		return err
	}
	return err

}

//beta. Не проверено в продукции

var (
	rabbitConn       *amqp.Connection
	rabbitCloseError chan *amqp.Error
)

func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}


func processMessage(body string, messageId string, replyTo string, userId  string, appId string, callBack string){

	atomic.AddInt32(&ops, 1)

	l := lua.NewState()
	lua.OpenLibraries(l)

	o := orm.NewOrm()
	o.Using("default")
	RegisterAPI(l, o)
	l.PushString(string(body))
	l.SetGlobal("body")
	l.PushString(messageId)
	l.SetGlobal("messageId")
	l.PushString(replyTo)
	l.SetGlobal("replyTo")
	l.PushString(userId)
	l.SetGlobal("userId")
	l.PushString(appId)
	l.SetGlobal("appId")

	if err := lua.DoString(l, callBack); err != nil {
		log.Println("error lua  " + err.Error())

	}
	atomic.AddInt32(&ops, -1)

}
func rabbitConnector(ampS string, uri string, L *lua.State, queue string, exchange string, routingKey string, durable bool, autoDelete bool, autoAck bool, prefetchCount int, callBack string) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError
		if rabbitErr != nil {
			log.Printf("Connecting to %s\n", uri)

			rabbitConn = connectToRabbitMQ(uri)
			rabbitCloseError = make(chan *amqp.Error)
			rabbitConn.NotifyClose(rabbitCloseError)

			ch, err := rabbitConn.Channel()
			if err != nil {
				log.Println("%s: %s", "Failed to open a channel", err)
				continue
			}

			if exchange != "" {

				err = ch.ExchangeDeclare(
					exchange,   // name
					"direct",   // type
					durable,    // durable
					autoDelete, // auto-deleted
					false,      // internal
					false,      // no-wait
					nil,        // arguments
				)

				if err != nil {
					log.Println("%s: %s", "Failed to declare an exchange", err)
					continue
				}
			}

			q, err := ch.QueueDeclare(
				queue,      // name
				durable,    // durable
				autoDelete, // delete when unused
				false,      // exclusive
				false,      // no-wait
				nil,        // arguments
			)
			if err != nil {
				log.Println("%s: %s", "Failed to declare a queue", err)
				continue
			}

			if exchange != "" {
				err = ch.QueueBind(
					q.Name,     // queue name
					routingKey, // routing key
					exchange,   // exchange
					false,
					nil)
			}

			err = ch.Qos(
				prefetchCount,	// prefetch count
				0,	// prefetch size
				false,	// global
			)
			if err != nil {
				log.Println("%s: %s", "Failed to set QoS", err)
				continue
			}

			log.Println("autoAck =", autoAck)
			log.Println("prefetchCount =", prefetchCount)

			msgs, err := ch.Consume(
				q.Name, // queue
				"",     // consumer
				autoAck,   // auto-ack
				false,  // exclusive
				false,  // no-local
				false,  // no-wait
				nil,    // args
			)

			if err != nil {
				log.Println("%s: %s", "Failed to register a consumer", err)
				continue
			}

			go func() {

				i := 0
				for d := range msgs {
					i++

					log.Println("Received a message", i)

					opsFinal := atomic.LoadInt32(&ops)

					if opsFinal > 100 {
						processMessage(string(d.Body),d.MessageId,d.ReplyTo,d.UserId,d.AppId,callBack)
						if !autoAck {
							log.Printf("Done")
							d.Ack(false)
						}
					} else {
						go func(d amqp.Delivery) {
							processMessage(string(d.Body),d.MessageId,d.ReplyTo,d.UserId,d.AppId,callBack)
							if !autoAck {
								log.Printf("Done")
								d.Ack(false)
							}
						}(d)
					}

				}

			}()

			log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

		}
	}
}

func TestReceive(L *lua.State, connStr string, queue string, exchange string, routingKey string, durable bool, autoDelete bool, autoAck bool, prefetchCount int, callBack string) error {

	flag.Parse()

	// create the rabbitmq error channel
	rabbitCloseError = make(chan *amqp.Error)

	// run the callback in a separate thread
	go rabbitConnector(*amqpUri, connStr, L, queue, exchange, routingKey, durable, autoDelete, autoAck, prefetchCount, callBack)

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback
	rabbitCloseError <- amqp.ErrClosed

	return nil

}

package main

import (
	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/slinky55/go-ws-chat/models"
	"github.com/slinky55/go-ws-chat/views"
	"golang.org/x/net/websocket"
	"sync"
	"time"
)

type PubSub struct {
	subscribers []chan models.Message
	mutex       sync.RWMutex
}

func (p *PubSub) Subscribe() <-chan models.Message {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	ch := make(chan models.Message)
	p.subscribers = append(p.subscribers, ch)

	return ch
}

func (p *PubSub) Publish(message models.Message) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, ch := range p.subscribers {
		ch <- message
	}
}

func main() {
	println("starting bailey...")

	e := echo.New()

	var ps PubSub
	var log []models.Message

	db, err := sqlx.Open("sqlite3", "./log.db")
	if err != nil {
		println(err.Error())
		return
	}

	db.MustExec(`CREATE TABLE IF NOT EXISTS log (text TEXT NOT NULL, author VARCHAR(15) DEFAULT NULL, timestamp TEXT NOT NULL)`)

	rows, err := db.Queryx("SELECT * FROM log")
	if err != nil {
		println("Failed to load message log")
		return
	}

	for rows.Next() {
		var text string
		var author string
		var timestamp string

		err = rows.Scan(&author, &text, &timestamp)
		if err != nil {
			println(err.Error())
			continue
		}

		log = append(log, models.Message{Text: author, Author: text, Timestamp: timestamp})
	}

	e.GET("/", func(c echo.Context) error {
		return Render(c, 200, views.Index(log))
	})

	e.GET("/subscribe", func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer func(ws *websocket.Conn) {
				err := ws.Close()
				if err != nil {
					c.Logger().Error(err.Error())
				}
			}(ws)

			incoming := ps.Subscribe()

			for {
				msg := <-incoming

				err := websocket.Message.Send(ws, `<div id="msgLog" hx-swap-oob="beforeend"><p>`+msg.Text+`</p></div>`)
				if err != nil {
					c.Logger().Error(err)
				}
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.POST("/chat", func(c echo.Context) error {
		var chat models.Message
		err := c.Bind(&chat)
		if err != nil {
			c.Logger().Error(err.Error())
			return err
		}

		chat.Timestamp = time.Now().Format(time.RFC3339)

		_, err = db.Exec("INSERT INTO log VALUES (?, ?, ?)", chat.Text, chat.Author, chat.Timestamp)
		if err != nil {
			c.Logger().Error(err.Error())
			return err
		}

		log = append(log, chat)

		ps.Publish(chat)

		return c.HTML(200, `<input type="text" name="msg" id="msg" required />`)
	})

	e.Logger.Fatal(e.Start(":80"))
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

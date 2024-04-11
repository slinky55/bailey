package main

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/slinky55/go-ws-chat/models"
	"github.com/slinky55/go-ws-chat/views"
	"golang.org/x/net/websocket"
	"sync"
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

	log := []models.Message{
		{Text: "Test!", Author: "Bob"},
		{Text: "Another test!", Author: "Jane"},
		{Text: "This is the 3rd test!", Author: "Joe"},
	}

	var ps PubSub

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
		log = append(log, chat)

		ps.Publish(chat)

		return c.HTML(200, `<input type="text" name="msg" id="msg" required />`)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

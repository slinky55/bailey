package main

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/slinky55/go-ws-chat/models"
	"github.com/slinky55/go-ws-chat/views"
)

func main() {
	println("starting bailey...")

	e := echo.New()

	log := []models.Message{
		{Text: "Test!", Author: "Bob"},
		{Text: "Another test!", Author: "Jane"},
		{Text: "This is the 3rd test!", Author: "Joe"},
	}

	e.GET("/", func(c echo.Context) error {
		return Render(c, 200, views.Index(log))
	})

	e.POST("/chat", func(c echo.Context) error {
		var chat models.Message
		err := c.Bind(&chat)
		if err != nil {
			c.Logger().Error(err.Error())
			return err
		}
		log = append(log, chat)
		if err != nil {
			return err
		}
		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

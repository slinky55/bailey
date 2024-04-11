package models

type Message struct {
	Text      string `form:"msg"`
	Author    string `form:"author"`
	Timestamp int64
}

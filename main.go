package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	uuid "github.com/nu7hatch/gouuid"

	"gopkg.in/olahol/melody.v1"
)

type wsMessage struct {
	Command string      `json:"command,omitempty"`
	Topic   string      `json:"topic"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	e := echo.New()
	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		m.HandleRequest(c.Response(), c.Request())
		return nil
	})

	m.HandleConnect(func(session *melody.Session) {
		//authToken := getAuthToken(session)
		userID := getUserID(session)
		connID, _ := uuid.NewV4()

		//session.Set("auth-token", authToken)
		session.Set("uid", userID)
		session.Set("connID", connID)
	})

	m.HandleMessage(func(session *melody.Session, msg []byte) {
		var message wsMessage
		json.Unmarshal(msg, &message)

		command := strings.ToLower(message.Command)
		topic := strings.ToLower(message.Topic)

		switch command {
		case "subscribe":
			addListener(topic, session)
		case "unsubscribe":
			removeListener(topic, session)
		case "publish":
			publish(topic, message.Data)
		default:
			fmt.Printf("%s not a valid command", command)
		}

		return
	})

	m.HandleDisconnect(func(session *melody.Session) {
		disconnectListener(session)
	})

	e.Logger.Fatal(e.Start(":5000"))
}

func getAuthToken(session *melody.Session) string {
	return getCookieValue(session.Request, "GRPTOK")
}

func getUserID(session *melody.Session) string {
	return getCookieValue(session.Request, "GRPUSR")
}

func getCookieValue(request *http.Request, cookieName string) string {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return cookie.Value
}

/*var socketRequest := interface{}
err := json.Unmarshal(msg, &socketRequest)
if err != nil {
	fmt.Println(err.Error())
	return
}

/*headers, err := json.Marshal(&socketRequest.Headers)
if err != nil {
	fmt.Println(err.Error())
	return
}

var headersInterface interface{}
if err = json.Unmarshal([]byte(headers), &headersInterface); err != nil {
	fmt.Println(err.Error())
	return
}

headersReflect := reflect.ValueOf(headersInterface)
headerKeys := headersReflect.MapKeys()

request := resty.R().SetBody(socketRequest.Body)
for _, headerKey := range headerKeys {
	key := fmt.Sprint(headerKey)
	headerValue := headersReflect.MapIndex(headerKey)

	value := fmt.Sprint(headerValue)
	request.Header.Add(key, value)
}

var resp *resty.Response

path := "http://localhost:1000" + socketRequest.Endpoint

switch method := socketRequest.Method; method {
case "GET":
	resp, err = request.Get(path)
case "POST":
	resp, err = request.Post(path)
case "PUT":
	resp, err = request.Put(path)
case "DELETE":
	resp, err = request.Delete(path)
default:
	return
}
if err != nil {
	fmt.Println(err.Error())
	return
}

responseHeaders, err := json.Marshal(resp.Header())
if err != nil {
	fmt.Println(err.Error())
	return
}

response, err := json.Marshal(responseStruct{
	Code:      resp.StatusCode(),
	Headers:   json.RawMessage(responseHeaders),
	Body:      json.RawMessage(resp.Body()),
	MessageID: socketRequest.MessageID,
})
if err != nil {
	fmt.Println(err.Error())
	return
}

s.Write(response)*/

package websocket

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/ilhasoft/wwcs/config"
	log "github.com/sirupsen/logrus"
)

// Client errors
var (
	// Register
	ErrorBlankFrom     = errors.New("unable to register: blank from")
	ErrorBlankCallback = errors.New("unable to register: blank callback")
	// Redirect
	ErrorNeedRegistration = errors.New("unable to redirect: id and url is blank")
	ErrorNoRedirects      = errors.New("unable to redirect: all redirects are desactivated")
)

// Client side data
type Client struct {
	ID       string
	Callback string
	Conn     *websocket.Conn
	Pool     *Pool
}

// ExternalPayload  data
type ExternalPayload struct {
	Type    string `json:"type"`
	To      string `json:"to,omitempty"`
	From    string `json:"from,omitempty"`
	Message Message
}

// SocketPayload data
type SocketPayload struct {
	Type     string  `json:"type"`
	From     string  `json:"from,omitempty"`
	Callback string  `json:"callback,omitempty"`
	Trigger  string  `json:"trigger,omitempty"`
	Message  Message `json:"message,omitempty"`
}

// Message data
type Message struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Text         string   `json:"text,omitempty"`
	URL          string   `json:"url,omitempty"`
	Caption      string   `json:"caption,omitempty"`
	FileName     string   `json:"filename,omitempty"`
	Latitude     string   `json:"latitude,omitempty"`
	Longitude    string   `json:"longitude,omitempty"`
	QuickReplies []string `json:"quick_replies,omitempty"`
}

// Sender message data
type Sender struct {
	Payload *ExternalPayload
	Client  *Client
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		log.Trace("Reading messages")
		socketPayload := SocketPayload{}
		err := c.Conn.ReadJSON(&socketPayload)
		if err != nil {
			if err.Error() != "websocket: close 1001 (going away)" {
				log.Error(err)
			}
			return
		}

		err = c.parsePayload(socketPayload)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (c *Client) parsePayload(payload SocketPayload) (err error) {
	switch payload.Type {
	case "register":
		err = c.Register(payload)
	case "message":
		_, err = c.Redirect(payload)
	}

	return
}

// Register register an user
func (c *Client) Register(payload SocketPayload) error {
	log.Tracef("Registering client %s", payload.From)
	if payload.From == "" {
		return ErrorBlankFrom
	}

	if payload.Callback == "" {
		return ErrorBlankCallback
	}

	c.ID = payload.From
	c.Callback = payload.Callback
	c.Pool.Register <- c

	// if has a trigger to start a flow, redirect it
	if payload.Trigger != "" {
		payload.Message.Text = payload.Trigger
		c.redirectToCallback(payload)
	}

	return nil
}

// Redirect message to the active redirects
func (c *Client) Redirect(payload SocketPayload) (int, error) {
	if c.ID == "" || c.Callback == "" {
		return 0, ErrorNeedRegistration
	}
	var redirects int

	config := config.Get.Websocket

	if config.RedirectToFrontend {
		c.redirectToFrontend(payload)
		redirects++
	}

	if config.RedirectToCallback {
		c.redirectToCallback(payload)
		redirects += 2
	}

	if redirects < 1 {
		return 0, ErrorNoRedirects
	}

	return redirects, nil
}

// redirectToCallback will send the message to the callback url provided on register
func (c *Client) redirectToCallback(payload SocketPayload) {
	log.Trace("Redirecting message to callback")
	form := url.Values{}
	form.Set("from", c.ID)
	form.Set("text", payload.Message.Text)

	req, _ := http.NewRequest("POST", c.Callback, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
	}
	log.Trace(res)
}

// redirectToFrontend will resend the message to the frontend
func (c *Client) redirectToFrontend(payload SocketPayload) {
	log.Trace("Redirecting message to frontend")
	external := &ExternalPayload{
		Message: Message{
			Text: payload.Message.Text,
		},
	}

	sender := Sender{
		Client:  c,
		Payload: external,
	}

	c.Pool.Sender <- sender
}

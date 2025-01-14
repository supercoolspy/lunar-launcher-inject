package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"time"
)

type ChromeDebugger struct {
	conn *websocket.Conn
}

func (d *ChromeDebugger) Close() error {
	return d.conn.Close()
}

func (d *ChromeDebugger) Send(method string, params map[string]any) error {
	return d.conn.WriteJSON(map[string]any{
		"id":     1,
		"method": method,
		"params": params,
	})
}

func ConnectDebugger(port int) (*ChromeDebugger, error) {
	url, err := GetWebsocketDebuggerUrl(port)
	if err != nil {
		return nil, err
	}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return &ChromeDebugger{
		conn: c,
	}, nil
}

func GetWebsocketDebuggerUrl(port int) (string, error) {
	var url string

	err := retry.Do(
		func() error {
			r, err := http.Get(fmt.Sprintf("http://localhost:%d/json/list", port))
			if err != nil {
				return err
			}
			defer r.Body.Close()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}

			var targets []struct {
				WebsocketUrl string `json:"webSocketDebuggerUrl"`
			}

			if err = json.Unmarshal(body, &targets); err != nil {
				return err
			} else if len(targets) == 0 {
				return errors.New("no debugging targets found")
			}

			url = targets[0].WebsocketUrl
			return nil
		},
		retry.Attempts(5),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(500*time.Millisecond),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		return "", err
	}

	return url, nil
}

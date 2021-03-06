package gosocketio

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gildas/golang-socketio/protocol"
	"github.com/gildas/golang-socketio/transport"
)

const (
	webSocketProtocol       = "ws://"
	webSocketSecureProtocol = "wss://"
	socketioUrl             = "/socket.io/?EIO=3&transport=websocket"
)

/**
Socket.io client representation
*/
type Client struct {
	methods
	Channel
}

/**
Get ws/wss url by host and port
*/
func GetUrl(host string, port int, secure bool) string {
	var prefix string
	if secure {
		prefix = webSocketSecureProtocol
	} else {
		prefix = webSocketProtocol
	}
	return prefix + host + ":" + strconv.Itoa(port) + socketioUrl
}

/**
connect to host and initialise socket.io protocol

The correct ws protocol url example:
ws://myserver.com/socket.io/?EIO=3&transport=websocket

You can use GetUrlByHost for generating correct url
*/
func Dial(url string, tr transport.Transport) (*Client, error) {
	c := &Client{}
	c.initChannel()
	c.initMethods()

	var err error
	c.conn, err = tr.Connect(url)
	if err != nil {
		return nil, err
	}

	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)
	c.On(OnDisconnection, func(channel *Channel, message interface{}) {
		c.Redial(url, tr)
	})

	return c, nil
}

/**
connect to host and initialise socket.io protocol

The correct ws protocol url example:
ws://myserver.com/socket.io/?EIO=3&transport=websocket

You can use GetUrlByHost for generating correct url
*/
func DialWithNamespace(url string, namespace string, tr transport.Transport) (*Client, error) {
	c := &Client{}
	c.initChannel()
	c.initMethods()

	var err error
	c.conn, err = tr.Connect(url)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("4%d%s", protocol.MessageTypeOpen, namespace)
	err = c.conn.WriteMessage(message)
	if err != nil {
		return nil, err
	}

	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)
	c.On(OnDisconnection, func(channel *Channel, message interface{}) {
		c.RedialWithNamespace(url, namespace, tr)
	})

	return c, nil
}

/**
Close client connection
*/
func (c *Client) Close() {
	closeChannel(&c.Channel, &c.methods)
}

/**
Redials
*/
func (c *Client) Redial(url string, tr transport.Transport) {
	var err error
	c.initChannel()
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			c.conn, err = tr.Connect(url)
			if err == nil {
				go inLoop(&c.Channel, &c.methods)
				go outLoop(&c.Channel, &c.methods)
				go pinger(&c.Channel)
				return
			}
		}
	}
}


/**
Redials
*/
func (c *Client) RedialWithNamespace(url string, namespace string, tr transport.Transport) {
	var err error
	c.initChannel()
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			c.conn, err = tr.Connect(url)
			if err == nil {
				message := fmt.Sprintf("4%d%s", protocol.MessageTypeOpen, namespace)
				err = c.conn.WriteMessage(message)
				if err != nil {
					continue
				}

				go inLoop(&c.Channel, &c.methods)
				go outLoop(&c.Channel, &c.methods)
				go pinger(&c.Channel)
				return
			}
		}
	}
}

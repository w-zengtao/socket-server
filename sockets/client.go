package sockets

import (
	"log"
	"time"
	
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 50 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client 是存在于服务端对连接的抽象描述 & 每一个连接都需要初始化一个 Client Instance
type Client struct {
	manager  *ClientManager
	conn     *websocket.Conn
	messages chan []byte
}

// Conn return physical connection
func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

// 这里暂时读消息只读心跳包
// 收到 Pong 的时候需要检查一下连接的地址是否依旧可用（Key过期）
// 暂时依旧使用 DB 来实现这个功能
func (c *Client) readMessageFromClient() {
	defer func() {
		c.manager.unregister <- c
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) // 从 conn读取 最多等待 pongWait 的时间 & 这个语句用来第一次
	c.conn.SetPongHandler(func(string) error {       // 这里设置 Pong 消息的处理器 & 如果没有收到 Pong 消息 那就会读到 Error
		//ip, _, _ := net.SplitHostPort(c.conn.RemoteAddr().String())
		//valid := verifyConn(ip)
		//if !valid {
		//	c.conn.Close() // 这里主动释放连接资源
		//	return errors.New("address not valid now")
		//}
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 这样的死循环一般来说开启一个 Goroutine
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}

}

// 写之前需要检查一下 Key 是否过期 过期需要切断连接
func (c *Client) writeMessageToClient() {
	ticker := time.NewTicker(pingPeriod) // 每隔 pingPeriod 触发一次 Ping 操作
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.messages:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) // 写入 Conn 的时间
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			// n := len(c.messages)
			// for i := 0; i < n; i++ {
			// 	w.Write([]byte{'\n'})
			// 	w.Write(<-c.messages)
			// }
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			log.Println("Sending Ping message to client ...")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Destory exports destory()
func (c *Client) Destory() {
	c.destory()
}

func (c *Client) destory() {
	delete(c.manager.clients, c)
	c.conn.Close()
	close(c.messages)
}

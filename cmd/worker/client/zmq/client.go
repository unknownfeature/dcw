package zmq

import (
	"errors"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/unknownfeature/dcw/cmd/util"
	"github.com/unknownfeature/dcw/cmd/worker/client"
	"log"
	"sync"
	"time"
)

type Config struct {
	Host                                 string
	Port                                 int
	Id                                   string
	MaxConnectAttempts                   int
	TimeToSleepBetweenConnectionAttempts time.Duration
}
type Client struct {
	socket    *zmq4.Socket
	config    Config
	connLock  *sync.Mutex
	connected bool
}

func NewZMQClient(config Config) (client.Client, error) {
	socket, err := zmq4.NewSocket(zmq4.DEALER)
	if err != nil {
		log.Printf("can't create socket %s", err.Error())
		return nil, err
	}
	for i := 0; i < config.MaxConnectAttempts; i++ {
		err = socket.Connect(fmt.Sprintf("tcp://%s:%d", config.Host, config.Port))
		if err != nil {
			log.Printf("can't connect to %s:%d %s", config.Host, config.Port, err)
			if i == config.MaxConnectAttempts-1 {
				return nil, err
			}
		} else {
			break
		}
		time.Sleep(config.TimeToSleepBetweenConnectionAttempts * time.Second)
	}
	connLock := sync.Mutex{}
	return &Client{socket, config, &connLock, true}, nil

}
func (c *Client) Call(req []byte) ([]byte, error) {
	c.connLock.Lock()
	_, err := c.socket.SendMessage(req)
	log.Println("sent")
	msg, err := c.socket.RecvMessage(0)
	log.Println("received")
	c.connLock.Unlock()
	if len(msg) > 0 {
		return util.SliceToByteSlice(msg), err
	}
	return nil, errors.New("invalid response")

}

func (c *Client) Close() error {
	c.connLock.Lock()
	err := c.Close()
	c.connected = false
	c.connLock.Unlock()
	return err
}

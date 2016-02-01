// Implementation of a MultiEchoServer. Students should write their code in this file.
// basicly ref: https://gist.github.com/drewolson/3950226

package p0

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type dataX struct {
	msg string
	err error
	c   *client
}

type client struct {
	id       int
	conn     net.Conn
	writeMsg chan string // The message to write to the network.
	reader   *bufio.Reader
	writer   *bufio.Writer
}

type multiEchoServer struct {
	// TODO: implement this!
	port     int
	listener net.Listener

	clients  []client
	join     chan net.Conn
	leave    chan dataX
	readChan chan dataX
	// writeChan chan dataX
	counts int
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!

	s := new(multiEchoServer)

	s.join = make(chan net.Conn, 100)
	s.leave = make(chan dataX, 100)
	s.readChan = make(chan dataX)
	// s.writeChan = make(chan dataX, 100)

	return s
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	// defer debug.PrintStack()
	mes.port = port

	ln, err := net.Listen("tcp", ":"+strconv.FormatInt(int64(port), 10))
	if err != nil {
		fmt.Println("Couldn't listen:", err)
		return err
	}

	mes.listener = ln

	go func() {

		for {
			// fmt.Println("Waiting for inbound connection")
			conn, err := mes.listener.Accept()
			if err != nil {
				// fmt.Println("Couldn't accept: ", err)
				continue
			}

			// mes.clientJoin(conn)

			// mes.counts++
			mes.join <- conn
		}
	}()

	go mes.run()

	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	mes.listener.Close()
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	return mes.counts
}

// TODO: add additional methods/functions below!

func (mes *multiEchoServer) clientJoin(c net.Conn) { //  conn net.Conn

	// fmt.Println("new client register")
	mes.counts++

	writeMsg := make(chan string, 100)

	mes.clients = append(mes.clients, client{
		conn:     c,
		reader:   bufio.NewReader(c),
		writer:   bufio.NewWriter(c),
		writeMsg: writeMsg,
	})

	cli := mes.clients[len(mes.clients)-1]
	// fmt.Println("client is mes.client", mes.clients)
	cli.run(mes)
}

func (mes *multiEchoServer) clientLeave(c *client) {
	mes.counts--
	for i, cli := range mes.clients {
		if cli == *c {
			// fmt.Println("client leaving ", cli)
			mes.clients = append(mes.clients[:i], mes.clients[i+1:]...)
		}
	}

}

func (mes *multiEchoServer) run() {
	for {
		select {
		case conn := <-mes.join:
			go mes.clientJoin(conn)

		case cmd := <-mes.leave:
			c := cmd.c
			go mes.clientLeave(c)

		case cmd := <-mes.readChan:
			msg, err, c := cmd.msg, cmd.err, cmd.c

			if err != nil {
				// mes.counts--
				// fmt.Println("mes client counts is ", mes.Count())
				mes.leave <- dataX{c: c}
				continue
			}

			for _, c := range mes.clients {
				//go c.writeToMsg(msg) // <- dataX{msg: msg}
				c := c
				go func() {
					c.writeMsg <- msg
				}()
			}

		}
	}
}

func (c *client) run(mes *multiEchoServer) {
	go c.read(mes)
	go c.write(mes)
}

func (c *client) read(mes *multiEchoServer) {
	for {
		msg, err := c.reader.ReadBytes('\n')

		if err != nil {
			// fmt.Println("client read err", err)
			//c.readMsg <- dataX{err: err,
			//	c: c}
			mes.readChan <- dataX{err: err,
				c: c}
			return
		}

		mes.readChan <- dataX{msg: string(msg)}
	}
}

func (c *client) write(mes *multiEchoServer) {
	for {
		for data := range c.writeMsg {

			_, err := c.writer.WriteString(data)
			// fmt.Println("data is ", data)

			if err != nil {
				fmt.Println("client write err", err)
				//continue
				return
			}

			c.writer.Flush()

		}
	}
}

package rabbitmq

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelPool struct {
	conn *RabbitMQConn
	pool chan *amqp.Channel
	size int
	mu   sync.Mutex
}

func NewChannelPool(conn *RabbitMQConn, size int) (*ChannelPool, error) {
	p := &ChannelPool{
		conn: conn,
		size: size,
		pool: make(chan *amqp.Channel, size),
	}

	for i := 0; i < size; i++ {
		ch, err := conn.NewChannel()

		if err != nil {
			return nil, err
		}

		p.pool <- ch
	}

	conn.pool = p

	return p, nil
}

func (p *ChannelPool) Get() (*amqp.Channel, error) {
	select {
	case ch := <-p.pool:
		if ch.IsClosed() {
			return p.conn.NewChannel()
		}

		return ch, nil
	default:
		return p.conn.NewChannel()
	}
}

func (p *ChannelPool) Put(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}

	select {
	case p.pool <- ch:
	default:
		ch.Close()

	}
}

func (p *ChannelPool) Rebuild() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.pool)

	p.pool = make(chan *amqp.Channel, p.size)

	for i := 0; i < p.size; i++ {
		ch, err := p.conn.NewChannel()

		if err == nil {
			p.pool <- ch
		}
	}
}

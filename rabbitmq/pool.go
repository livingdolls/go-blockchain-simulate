package rabbitmq

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelPool struct {
	conn   *RabbitMQConn
	pool   chan *amqp.Channel
	size   int
	mu     sync.RWMutex
	closed bool
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
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, amqp.ErrClosed
	}

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

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		ch.Close()
		return
	}

	select {
	case p.pool <- ch:
	default:
		ch.Close()

	}
}

func (p *ChannelPool) Rebuild() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldPool := p.pool

	p.pool = make(chan *amqp.Channel, p.size)

	for i := 0; i < p.size; i++ {
		ch, err := p.conn.NewChannel()

		if err != nil {
			continue
		}

		p.pool <- ch
	}

	go func() {
		for {
			select {
			case ch := <-oldPool:
				ch.Close()
			default:
				return
			}
		}
	}()

	return nil
}

func (p *ChannelPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	close(p.pool)

	for ch := range p.pool {
		ch.Close()
	}
}

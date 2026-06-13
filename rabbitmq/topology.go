package rabbitmq

import (
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

func (c *RabbitMQConn) DeclareQueue(q models.QueueDef) error {
	ch, err := c.NewChannel()
	if err != nil {
		return err
	}

	defer ch.Close()

	_, err = ch.QueueDeclare(q.Name, q.Durable, q.AutoDelete, false, false, nil)

	if err == nil {
		// Setup success tidak di-log - ini terjadi sekali saat startup,
		// bukan operasional. Error sudah di-return via err untuk caller.
		// Kalau perlu debug, pakai AMQP server log atau inspect connection.
		c.queues = append(c.queues, q)
	}

	return err
}

func (c *RabbitMQConn) DeclareExchange(e models.ExchangeDef) error {
	ch, err := c.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(e.Name, e.Kind, e.Durable, false, false, false, nil)

	if err == nil {
		c.exchanges = append(c.exchanges, e)
	}
	return err
}

func (c *RabbitMQConn) Bind(b models.BindDef) error {
	ch, err := c.NewChannel()
	if err != nil {
		return err
	}

	defer ch.Close()

	err = ch.QueueBind(b.Queue, b.RoutingKey, b.Exchange, false, nil)

	if err == nil {
		c.binds = append(c.binds, b)
	}

	return err
}

func (c *RabbitMQConn) restoreTopology() {
	ch, _ := c.NewChannel()
	defer ch.Close()

	for _, e := range c.exchanges {
		ch.ExchangeDeclare(e.Name, e.Kind, e.Durable, false, false, false, nil)
	}

	for _, q := range c.queues {
		ch.QueueDeclare(q.Name, q.Durable, q.AutoDelete, false, false, nil)
	}

	for _, b := range c.binds {
		ch.QueueBind(b.Queue, b.RoutingKey, b.Exchange, false, nil)
	}
}

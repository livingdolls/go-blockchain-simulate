package rabbitmq

import (
	"log"

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
		log.Printf("[RABBITMQ] Declared queue: %s", q.Name)
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
		log.Printf("[RABBITMQ] Declared exchange: %s", e.Name)
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
		log.Printf("[RABBITMQ] Bound queue %s to exchange %s with routing key %s", b.Queue, b.Exchange, b.RoutingKey)
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

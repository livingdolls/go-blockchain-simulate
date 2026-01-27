package models

type QueueDef struct {
	Name       string
	Durable    bool
	AutoDelete bool
}

type ExchangeDef struct {
	Name    string
	Kind    string
	Durable bool
}

type BindDef struct {
	Queue      string
	Exchange   string
	RoutingKey string
}

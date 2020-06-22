package pubsub

import "go-pubsub-ws/pkg/websocket"

type Publication struct {
	origin      websocket.WSConn
	namespace   string
	destination string
	data        interface{}
}

func CreatePublication(origin websocket.WSConn, destination, publicationType string, data interface{}) *Publication {
	return &Publication{origin, destination, publicationType, data}
}

func (p *Publication) Origin() websocket.WSConn {
	return p.origin
}

func (p *Publication) Destination() string {
	return p.destination
}

func (p *Publication) SetDestination(destination string) {
	p.destination = destination
}

func (p *Publication) Namespace() string {
	return p.namespace
}

func (p *Publication) SetNamespace(namespace string) {
	p.namespace = namespace
}

func (p *Publication) Data() interface{} {
	return p.data
}

func (p *Publication) SetData(data interface{}) {
	p.data = data
}

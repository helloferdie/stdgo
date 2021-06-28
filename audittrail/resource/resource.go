package resource

import "github.com/helloferdie/stdgo/librabbitmq"

// Get -
func Get() *librabbitmq.Resource {
	res := new(librabbitmq.Resource)
	res.Queue = librabbitmq.Queue{
		Name:      "audittrail",
		Durable:   true,
		Exclusive: false,
		NoWait:    false,
	}
	return res
}

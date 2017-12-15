package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
)

type HTTPRequestListener struct {
	Emitter func(ctx context.Context, c *crew.Crew, message interface{})
}

func (l *HTTPRequestListener) Handle(ctx context.Context, c *crew.Crew, message interface{}) {
	log.Printf("HTTPRequestListener Handle %s", JS(message))

	var to interface{}
	{
		bss, err := core.Match(nil, Dwimjs(`{"replyTo":{"mid":"?mid"}}`), message, core.NewBindings())
		if err == nil {
			if 0 < len(bss) {
				to, _ = bss[0]["?mid"]
			}
		}
	}

	// Sorry.
	js, err := json.Marshal(&message)
	if err != nil {
		// ToDo: Better than this.
		log.Printf("HTTPRequestListener Marshal error %v", err)
		return
	}

	var r HTTPRequest
	if err = json.Unmarshal(js, &r); err != nil {
		// ToDo: Better than this.
		log.Printf("HTTPRequestListener Marshal error %v", err)
		return
	}

	err = r.Do(ctx, func(ctx context.Context, resp *HTTPResponse) error {
		js, err := json.Marshal(&resp)
		if err != nil {
			log.Printf("HTTPRequestListener Marshal error %s", err)
			return err
		}
		var message map[string]interface{}
		if err = json.Unmarshal(js, &message); err != nil {
			log.Printf("HTTPRequestListener Unmarshal error %s", err)
			return err
		}
		if to != nil {
			message["to"] = map[string]interface{}{
				"mid": to,
			}
		}
		l.Emitter(ctx, c, message)
		return nil
	})

	if err != nil {
		// ToDo: Better than this.
		log.Printf("HTTPRequestListener request error %v", err)
		return
	}

}

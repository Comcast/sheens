package main

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/Comcast/sheens/util/testutil"

	"github.com/Comcast/sheens/core"
)

func (s *Service) toHTTP(ctx context.Context, msg interface{}) error {
	m, is := msg.(map[string]interface{})
	if !is {
		return fmt.Errorf("HTTP error %ts (%T) isn't a %T", JS(msg), msg, m)
	}

	msg, have := m["request"]
	if !have {
		return fmt.Errorf("HTTP error no 'request' in %s", JS(m))
	}

	var replyTo interface{}

	bss, err := core.Match(nil, Dwimjs(`{"replyTo":"?mid"}`), msg, core.NewBindings())
	if err == nil {
		if 0 < len(bss) {
			replyTo, _ = bss[0]["?mid"]
		}
	}

	var r HTTPRequest
	{
		// Sorry.
		js, err := json.Marshal(&msg)
		if err != nil {
			return fmt.Errorf("Service toHTTP Marshal error %s", err)
			return err
		}
		if err = json.Unmarshal(js, &r); err != nil {
			return fmt.Errorf("Service toHTTP Unmarshal error %s", err)
		}
	}

	err = r.Do(ctx, func(ctx context.Context, resp *HTTPResponse) error {
		// Again: sorry.
		js, err := json.Marshal(&resp)
		if err != nil {
			return fmt.Errorf("Service toHTTP result Marshal error %s", err)
		}
		var msg map[string]interface{}
		if err = json.Unmarshal(js, &msg); err != nil {
			return fmt.Errorf("Service toHTTP result Unmarshal error %s", err)
		}
		if replyTo != nil {
			msg["to"] = replyTo
		}

		if _, err := s.Process(ctx, msg, s.ProcessCtl); err != nil {
			return fmt.Errorf("Service toHTTP Process error %s", err)
		}

		return nil
	})

	return err
}

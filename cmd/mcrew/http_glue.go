/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Comcast/sheens/match"
	. "github.com/Comcast/sheens/util/testutil"
)

func (s *Service) toHTTP(ctx context.Context, msg interface{}) error {
	m, is := msg.(map[string]interface{})
	if !is {
		return fmt.Errorf("HTTP error %s (%T) isn't a %T", JS(msg), msg, m)
	}

	msg, have := m["request"]
	if !have {
		return fmt.Errorf("HTTP error no 'request' in %s", JS(m))
	}

	var replyTo interface{}

	bss, err := match.Match(Dwimjs(`{"replyTo":"?mid"}`), msg, match.NewBindings())
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
		resp.From = "http" // me

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

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
	"fmt"
	"time"

	testutils "github.com/Comcast/sheens/util/testutil"
)

// toTimers is a tedious method that attempts to interpret msg as
// either a 'makeTimer' or 'deleteTimer' request.
func (s *Service) toTimers(ctx context.Context, msg interface{}) error {
	m, is := msg.(map[string]interface{})
	if !is {
		return fmt.Errorf("%s (%T) isn't a %T", testutils.JS(msg), msg, m)
	}

	if v, have := m["makeTimer"]; have {

		if m, is = v.(map[string]interface{}); !is {
			return fmt.Errorf("makeTimer: %s (%T) isn't a %T", testutils.JS(msg), msg, m)
		}

		var id string
		if x, have := m["id"]; have {
			if id, is = x.(string); !is {
				return fmt.Errorf("id %s (%T) isn't a %T", testutils.JS(x), x, id)
			}
		}

		var d time.Duration
		var err error
		if x, have := m["in"]; have {
			str, is := x.(string)
			if !is {
				return fmt.Errorf("'in' %s (%T) isn't a %T", testutils.JS(x), x, id)
			}
			if d, err = time.ParseDuration(str); err != nil {
				return fmt.Errorf("bad duration '%s': %s", str, err)
			}
		} else if x, have := m["at"]; have {
			str, is := x.(string)
			if !is {
				return fmt.Errorf("'in' %s (%T) isn't a %T", testutils.JS(x), x, id)
			}
			var t time.Time
			t, err = time.Parse(time.RFC3339, str)
			if err != nil {
				return fmt.Errorf("bad RFC3339 time '%s': %s", str, err)
			}
			d = t.Sub(time.Now())
		} else {
			return fmt.Errorf("no 'at' or 'in' in %s", testutils.JS(m))
		}

		msg, have := m["message"]
		if !have {
			return fmt.Errorf("no 'message' in %s", testutils.JS(m))
		}

		if err = s.timers.Add(ctx, id, msg, d); err != nil {
			return fmt.Errorf("error for makeTimer %s: %s", id, err)
		}
	} else if x, have := m["deleteTimer"]; have {
		id, is := x.(string)
		if !is {
			return fmt.Errorf("id %s (%T) isn't a %T", testutils.JS(x), x, id)
		}
		if err := s.timers.Rem(ctx, id); err != nil {
			return fmt.Errorf("error for deleteTimer %s: %s", id, err)
		}
	} else {
		return fmt.Errorf("no 'makeTimer' or 'deleteTimer' in %s", testutils.JS(msg))
	}

	return nil
}

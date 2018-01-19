package main

import (
	"context"
	"fmt"
	"time"
)

// toTimers is a tedious method that attempts to interpret msg as
// either a 'makeTimer' or 'deleteTimer' request.
func (s *Service) toTimers(ctx context.Context, msg interface{}) error {
	m, is := msg.(map[string]interface{})
	if !is {
		return fmt.Errorf("%s (%T) isn't a %T", JS(msg), msg, m)
	}

	if v, have := m["makeTimer"]; have {

		if m, is = v.(map[string]interface{}); !is {
			return fmt.Errorf("makeTimer: %s (%T) isn't a %T", JS(msg), msg, m)
		}

		var id string
		if x, have := m["id"]; have {
			if id, is = x.(string); !is {
				return fmt.Errorf("id %s (%T) isn't a %T", JS(x), x, id)
			}
		}

		var d time.Duration
		var err error
		if x, have := m["in"]; have {
			str, is := x.(string)
			if !is {
				return fmt.Errorf("'in' %s (%T) isn't a %T", JS(x), x, id)
			}
			if d, err = time.ParseDuration(str); err != nil {
				return fmt.Errorf("bad duration '%s': %s", str, err)
			}
		} else if x, have := m["at"]; have {
			str, is := x.(string)
			if !is {
				return fmt.Errorf("'in' %s (%T) isn't a %T", JS(x), x, id)
			}
			var t time.Time
			t, err = time.Parse(time.RFC3339, str)
			if err != nil {
				return fmt.Errorf("bad RFC3339 time '%s': %s", str, err)
			}
			d = t.Sub(time.Now())
		} else {
			return fmt.Errorf("no 'at' or 'in' in %s", JS(m))
		}

		msg, have := m["message"]
		if !have {
			return fmt.Errorf("no 'message' in %s", JS(m))
		}

		if err = s.timers.Add(ctx, id, msg, d); err != nil {
			return fmt.Errorf("error for makeTimer %s: %s", id, err)
		}
	} else if x, have := m["deleteTimer"]; have {
		id, is := x.(string)
		if !is {
			return fmt.Errorf("id %s (%T) isn't a %T", JS(x), x, id)
		}
		if err := s.timers.Rem(ctx, id); err != nil {
			return fmt.Errorf("error for deleteTimer %s: %s", id, err)
		}
	} else {
		return fmt.Errorf("no 'makeTimer' or 'deleteTimer' in %s", JS(msg))
	}

	return nil
}

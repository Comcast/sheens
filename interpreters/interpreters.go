package interpreters

import (
	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/interpreters/ecmascript"
	"github.com/Comcast/sheens/interpreters/noop"
)

func Standard() core.InterpretersMap {
	is := core.NewInterpretersMap()

	es := ecmascript.NewInterpreter()
	is["ecmascript"] = es
	is["ecmascript-5.1"] = es
	is[""] = es // Default

	ext := ecmascript.NewInterpreter()
	ext.Extended = true
	is["ecmascript-ext"] = ext
	is["ecmascript-5.1-ext"] = ext

	is["noop"] = noop.NewInterpreter()

	// For backwards compatibility
	is["goja"] = ext

	return is
}

package coolbee

import (
	"reflect"
)

type ReturnHandler func(Context, []reflect.Value)

func defaultReturnHandler() ReturnHandler {
	return func(ctx Context, vals []reflect.Value) {
		
	}
}


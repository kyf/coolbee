package coolbee

import (
	"log"
	"io"
	"reflect"

	"github.com/codegangsta/inject"
)

type Coolbee struct {
	inject.Injector
	handlers []Handler
	action   Handler
	logger   *log.Logger
	endHandlers []Handler
}

func New(out io.Writer) *Coolbee {
	m := &Coolbee{Injector: inject.New(), action: func() {}, logger: log.New(out, "[coolbee] ", log.LstdFlags)}
	m.Map(m.logger)
	m.Map(defaultReturnHandler())
	return m
}


func (m *Coolbee) Handlers(handlers ...Handler) {
	m.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		m.Use(handler)
	}
}

func (m *Coolbee) End(handler Handler) {
	validateHandler(handler)
	m.endHandlers = append(m.endHandlers, handler)
}



func (m *Coolbee) Action(handler Handler) {
	validateHandler(handler)
	m.action = handler
}


func (m *Coolbee) Use(handler Handler) {
	validateHandler(handler)

	m.handlers = append(m.handlers, handler)
}


func (m *Coolbee) Do() {
	m.createContext().run()
}


func (m *Coolbee) Run() {
	m.Do()
}

func (m *Coolbee) createContext() *context {
	c := &context{inject.New(), m.handlers, m.action, 0, m.endHandlers}
	c.SetParent(m)
	c.MapTo(c, (*Context)(nil))
	return c
}


type ClassicCoolbee struct {
	*Coolbee
	Router
}


func Classic(out io.Writer) *ClassicCoolbee {
	r := NewRouter()
	m := New(out)
	m.Use(Logger())
	m.Use(Recovery())
	m.MapTo(r, (*Router)(nil))
	m.Action(r.Handle)
	return &ClassicCoolbee{m, r}
}

type Handler interface{}

func validateHandler(handler Handler) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic("coolbee handler must be a callable func")
	}
}


type Context interface {
	inject.Injector
	Next()
	getEndHandlers()[]Handler
}

type context struct {
	inject.Injector
	handlers []Handler
	action   Handler
	index    int
	endHandlers []Handler
}

func (c *context) handler() Handler {
	if c.index < len(c.handlers) {
		return c.handlers[c.index]
	}
	if c.index == len(c.handlers) {
		return c.action
	}
	panic("invalid index for context handler")
}

func (c *context) Next() {
	c.index += 1
	c.run()
}

func (c *context) getEndHandlers() []Handler {
	return c.endHandlers
}


func (c *context) run() {
	for c.index <= len(c.handlers) {
		_, err := c.Invoke(c.handler())
		if err != nil {
			panic(err)
		}
		c.index += 1
	}
}

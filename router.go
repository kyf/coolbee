package coolbee

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"io/ioutil"
	"log"
	"sync"
)

type RespBody  interface{}

type Router interface {
	Group(string, func(Router), ...Handler)

	Get(string, ...Handler) Route

	AddRoute(string, string, ...Handler) Route

	Handle(Context, *log.Logger)
}

type router struct {
	routes     []*route
	groups     []group
	routesLock sync.RWMutex
}

type group struct {
	pattern  string
	handlers []Handler
}

func NewRouter() Router {
	return &router{groups: make([]group, 0)}
}

func (r *router) Group(pattern string, fn func(Router), h ...Handler) {
	r.groups = append(r.groups, group{pattern, h})
	fn(r)
	r.groups = r.groups[:len(r.groups)-1]
}

func (r *router) Get(pattern string, h ...Handler) Route {
	return r.addRoute("GET", pattern, h)
}

func (r *router) AddRoute(method, pattern string, h ...Handler) Route {
	return r.addRoute(method, pattern, h)
}

func (r *router) Handle(context Context, logger *log.Logger) {
	for _, route := range r.getRoutes() {
		resp, err := http.Get(route.Pattern());
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Println(err)
			}else{
				context.MapTo(string(body), (*RespBody)(nil))
				route.Handle(context)
				endHandlers := context.getEndHandlers();
				for _, h := range endHandlers {
					context.Invoke(h)
				}
			}
		}else{
			logger.Println(err)
		}
	}
}

func (r *router) addRoute(method string, pattern string, handlers []Handler) *route {
	if len(r.groups) > 0 {
		groupPattern := ""
		h := make([]Handler, 0)
		for _, g := range r.groups {
			groupPattern += g.pattern
			h = append(h, g.handlers...)
		}

		pattern = groupPattern + pattern
		h = append(h, handlers...)
		handlers = h
	}

	route := newRoute(method, pattern, handlers)
	route.Validate()
	r.appendRoute(route)
	return route
}

func (r *router) appendRoute(rt *route) {
	r.routesLock.Lock()
	defer r.routesLock.Unlock()
	r.routes = append(r.routes, rt)
}

func (r *router) getRoutes() []*route {
	r.routesLock.RLock()
	defer r.routesLock.RUnlock()
	return r.routes[:]
}

func (r *router) findRoute(name string) *route {
	for _, route := range r.getRoutes() {
		if route.name == name {
			return route
		}
	}

	return nil
}


type Route interface {

	Name(string)

	GetName() string

	Pattern() string

	Method() string
}

type route struct {
	method   string
	regex    *regexp.Regexp
	handlers []Handler
	pattern  string
	name     string
}

var routeReg1 = regexp.MustCompile(`:[^/#?()\.\\]+`)
var routeReg2 = regexp.MustCompile(`\*\*`)

func newRoute(method string, pattern string, handlers []Handler) *route {
	route := route{method, nil, handlers, pattern, ""}
	pattern = routeReg1.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	var index int
	pattern = routeReg2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	route.regex = regexp.MustCompile(pattern)
	return &route
}

func (r *route) Validate() {
	for _, handler := range r.handlers {
		validateHandler(handler)
	}
}

func (r *route) Handle(c Context) {
	context := &routeContext{c, 0, r.handlers}
	c.MapTo(context, (*Context)(nil))
	c.MapTo(r, (*Route)(nil))
	context.run()
}


func (r *route) Name(name string) {
	r.name = name
}

func (r *route) GetName() string {
	return r.name
}

func (r *route) Pattern() string {
	return r.pattern
}

func (r *route) Method() string {
	return r.method
}

type routeContext struct {
	Context
	index    int
	handlers []Handler
}

func (r *routeContext) Next() {
	r.index += 1
	r.run()
}

func (r *routeContext) run() {
	for r.index < len(r.handlers) {
		handler := r.handlers[r.index]
		vals, err := r.Invoke(handler)
		if err != nil {
			panic(err)
		}
		r.index += 1

		if len(vals) > 0 {
			ev := r.Get(reflect.TypeOf(ReturnHandler(nil)))
			handleReturn := ev.Interface().(ReturnHandler)
			handleReturn(r, vals)
		}
	}
}

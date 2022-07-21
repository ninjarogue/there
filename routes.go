package there

import (
	"errors"
	"fmt"
	"strings"
)

type RouteGroup struct {
	*Router
	prefix string
}

func (group RouteGroup) Group(prefix string) *RouteGroup {

	prefix = strings.TrimPrefix(prefix, "/")

	group.assert(len(prefix) > 1, "route group needs to have at least one symbol")

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	group.prefix += prefix
	return &group
}


func NewRouteGroup(router *Router, route string) *RouteGroup {

	router.assert(route != "", "route \""+route+"\" must not be empty")

	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	if !strings.HasSuffix(route, "/") {
		route += "/"
	}

	return &RouteGroup{
		Router: router,
		prefix: route,
	}
}

type Endpoint func(request Request) Response

//Route adds attributes to an Endpoint func
type Route struct {
	Endpoint    Endpoint
	Methods     []string
	Path        Path
	Middlewares []Middleware
	stringPath    string
}

//OverlapsWith checks if an Route somehow overlaps with another container. For this to be true, the path and at least one method must equal
func (e Route) OverlapsWith(toCompare Route) bool {
	if !e.Path.Equals(toCompare.Path) {
		return false
	}
	return CheckArraysOverlap(e.Methods, toCompare.Methods)
}

func (e Route) ToString() string {
	r := fmt.Sprint(e.Methods, " ", e.Path.ToString())
	if e.Path.ignoreCase {
		r += " *IgnoreCase"
	}
	return r
}

func (group *RouteGroup) Handle(path string, endpoint Endpoint, methods ...string) *RouteRouteGroupBuilder {
	switch {
		case path[0] == '/' && len(path) > 1:
			path = strings.TrimPrefix(path, "/")
			path = strings.TrimSuffix(path, "/")
			fallthrough
		case path != "/":
			path = group.prefix + path
			fallthrough
		case path == "":
			path = "/"
	}

	if group.base.methodTrees == nil {
		group.base.methodTrees = make(map[string]methodTree)
	}

	route := &Route{
		endpoint,
		methods,
		ConstructPath(path, false),
		make([]Middleware, 0),
		path,
	}
	group.base.AddRoute(route)

	return &RouteRouteGroupBuilder{
		route,
		group,
	}
}

type RouteRouteGroupBuilder struct {
	*Route
	*RouteGroup
}

func (group *RouteGroup) Get(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodGet)
}

func (group *RouteGroup) Post(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodPost)
}

func (group *RouteGroup) Patch(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodPatch)
}

func (group *RouteGroup) Delete(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodDelete)
}

func (group *RouteGroup) Connect(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodConnect)
}

func (group *RouteGroup) Head(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodHead)
}

func (group *RouteGroup) Trace(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodTrace)
}

func (group *RouteGroup) Put(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodPut)
}

func (group *RouteGroup) Options(route string, endpoint Endpoint) *RouteRouteGroupBuilder {
	return group.Handle(route, endpoint, MethodOptions)
}

//With adds a middleware to the handler the method is called on
func (group *RouteRouteGroupBuilder) With(middleware Middleware) *RouteRouteGroupBuilder {
	group.Middlewares = append(group.Middlewares, middleware)
	return group
}

func (group *RouteRouteGroupBuilder) IgnoreCase() *RouteRouteGroupBuilder {
	// cancel if already ignore case
	if group.Route.Path.ignoreCase {
		return group
	}

	// delete route
	group.routes.RemoveRoute(group.Route)

	group.Route.Path.ignoreCase = true
	group.routes.AddRoute(group.Route, group.Router)

	return group
}

type routeManager []*Route

func (r *routeManager) FindOverlappingRoute(routeToCheck *Route) *Route {
	for _, toCompare := range *r {
		if toCompare.OverlapsWith(*routeToCheck) {
			return toCompare
		}
	}
	return nil
}

func (r *routeManager) AddRoute(routeToAdd *Route, router *Router) *Route {
	overlapsWith := r.FindOverlappingRoute(routeToAdd)
	if overlapsWith != nil {
		router.addError(errors.New("the route \""+routeToAdd.ToString()+"\" overlaps with the existing route \""+overlapsWith.ToString()+"\""))
		return nil
	}
	*r = append(*r, routeToAdd)
	return routeToAdd
}

func (r *routeManager) RemoveRoute(toRemove *Route) {
	for i, container := range *r {
		if container.Path.Equals(toRemove.Path) {
			*r = append((*r)[:i], (*r)[i+1:]...)
		}
	}
}

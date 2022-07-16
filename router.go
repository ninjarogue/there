package there

import (
	"errors"
	"fmt"
	"net/http"
)

type Router struct {
	*RouteGroup
	Configuration *RouterConfiguration
	Server        *http.Server

	assertionErrors

	globalMiddlewares []Middleware

	//routes is a list of Routes which checks for duplicate entries
	//on insert.
	routes routeManager
	methodTrees []MethodTree
}

func NewRouter() *Router {
	r := &Router{
		globalMiddlewares: make([]Middleware, 0),
		routes:            make([]*Route, 0),
		Server:            &http.Server{},
		Configuration: &RouterConfiguration{
			RouteNotFoundHandler: func(request Request) Response {
				return Error(StatusNotFound, errors.New("could not find route "+request.Method+" "+request.Request.URL.Path))
			},
		},
	}
	r.Server.Handler = r
	r.RouteGroup = NewRouteGroup(r, "/")
	return r
}

type Port uint16

func (p Port) ToAddr() string {
	return fmt.Sprintf(":%d", p)
}

func (router *Router) Listen(port Port) error {
	err := router.HasError()
	if err != nil {
		return err
	}
	router.Server.Addr = port.ToAddr()
	return router.Server.ListenAndServe()
}

func (router *Router) ListenToTLS(port Port, certFile, keyFile string) error {
	err := router.HasError()
	if err != nil {
		return err
	}
	router.Server.Addr = port.ToAddr()
	return router.Server.ListenAndServeTLS(certFile, keyFile)
}

//Use registers a Middleware
func (router *Router) Use(middleware Middleware) *Router {
	router.globalMiddlewares = append(router.globalMiddlewares, middleware)
	return router
}

//RouterConfiguration is a straightforward place to override default behavior of the router
type RouterConfiguration struct {
	//RouteNotFoundHandler gets invoked, when the specified URL and method have no handlers
	RouteNotFoundHandler Endpoint
}

type assertionErrors []error

func (a *assertionErrors) HasError() error {
	if len(*a) == 0 {
		return nil
	}
	var err error
	err, *a = (*a)[0], (*a)[1:]
	return err
}

func (a *assertionErrors) assert(condition bool, errorString string) {
	if !condition {
		*a = append(*a, errors.New(errorString))
	}
}

func (a *assertionErrors) addError(err error) {
	*a = append(*a, err)
}

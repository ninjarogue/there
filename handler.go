package there

import (
	"net/http"
)

func (router *Router) ServeHTTP(rw http.ResponseWriter, request *http.Request) {

	httpRequest := NewHttpRequest(rw, request)
	var middlewares = make([]Middleware, 0)
	middlewares = append(middlewares, router.globalMiddlewares...)

	var endpoint Endpoint = nil

	// We fetch the route along with the routeParams (if any).
	r, routeParams, _ := router.base.LookUp(request.Method, request.URL.Path)

	if r != nil {
		endpoint = r.endpoint
		middlewares = append(middlewares, r.middlewares...)
		routeParamReader := RouteParamReader(routeParams)
		httpRequest.RouteParams = &routeParamReader
	}

	if endpoint == nil {
		endpoint = router.Configuration.RouteNotFoundHandler
	}

	var next Response = ResponseFunc(func(rw http.ResponseWriter, r *http.Request) {
		endpoint(httpRequest).ServeHTTP(rw, r)
	})
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		next = middleware(httpRequest, next)
	}
	next.ServeHTTP(rw, request)
}

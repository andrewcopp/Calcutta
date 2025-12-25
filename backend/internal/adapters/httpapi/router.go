package httpapi

import "net/http"

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type Router struct{}

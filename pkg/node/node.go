package node

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

const apiPrefix = "/node"

// Node is the Node to let the controller to control the node.
type Node struct {
	name   string
	addr   string
	s      *http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

// NewNode creates the node with given address
func NewNode(name string, addr string) *Node {
	n := &Node{
		name: name,
		addr: addr,
	}

	n.ctx, n.cancel = context.WithCancel(context.Background())
	return n
}

// Run runs the Node API server
func (n *Node) Run() error {
	n.s = &http.Server{
		Addr:    n.addr,
		Handler: n.createHandler(),
	}
	return n.s.ListenAndServe()
}

// Close closes the Node.
func (n *Node) Close() {
	n.cancel()
	if n.s != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		n.s.Shutdown(ctx)
		cancel()
	}
}

func (n *Node) createHandler() http.Handler {
	engine := negroni.New()
	recover := negroni.NewRecovery()
	engine.Use(recover)

	router := mux.NewRouter()
	subRouter := n.createRouter()
	router.PathPrefix(apiPrefix).Handler(
		negroni.New(negroni.Wrap(subRouter)),
	)

	engine.UseHandler(router)
	return engine
}

func (n *Node) createRouter() *mux.Router {
	rd := render.New(render.Options{
		IndentJSON: true,
	})

	router := mux.NewRouter().PathPrefix(apiPrefix).Subrouter()

	nemesisHandler := newNemesisHandler(n, rd)
	router.HandleFunc("/nemesis/{name}/setup", nemesisHandler.SetUp).Methods("POST")
	router.HandleFunc("/nemesis/{name}/teardown", nemesisHandler.TearDown).Methods("POST")
	router.HandleFunc("/nemesis/{name}/invoke", nemesisHandler.Invoke).Methods("POST")
	
	dbHandler := newDBHanlder(n, rd) 
	router.HandleFunc("/db/{name}/setup", dbHandler.SetUp).Methods("POST")
	router.HandleFunc("/db/{name}/teardown", dbHandler.TearDown).Methods("POST")
	
	return router
}

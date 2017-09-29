package node

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

const apiPrefix = "/agent"

// Agent is the agent to let the controller to control the node.
type Agent struct {
	addr   string
	s      *http.Server
	ctx    context.Context
	cancel context.CancelFunc

	dbLock      sync.Mutex
	nemesisLock sync.Mutex
}

// NewAgent creates the agent with given address
func NewAgent(addr string) *Agent {
	agent := &Agent{
		addr: addr,
	}

	agent.ctx, agent.cancel = context.WithCancel(context.Background())
	return agent
}

// Run runs the agent API server
func (agent *Agent) Run() error {
	agent.s = &http.Server{
		Addr:    agent.addr,
		Handler: agent.createHandler(),
	}
	return agent.s.ListenAndServe()
}

// Close closes the agent.
func (agent *Agent) Close() {
	agent.cancel()
	if agent.s != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		agent.s.Shutdown(ctx)
		cancel()
	}
}

func (agent *Agent) createHandler() http.Handler {
	engine := negroni.New()
	recover := negroni.NewRecovery()
	engine.Use(recover)

	router := mux.NewRouter()
	subRouter := agent.createRouter()
	router.PathPrefix(apiPrefix).Handler(
		negroni.New(negroni.Wrap(subRouter)),
	)

	engine.UseHandler(router)
	return engine
}

func (agent *Agent) createRouter() *mux.Router {
	rd := render.New(render.Options{
		IndentJSON: true,
	})

	router := mux.NewRouter().PathPrefix(apiPrefix).Subrouter()

	nemesisHandler := newNemesisHandler(agent, rd)
	// router.HandleFunc("/nemesis/{name}/setup", nemesisHandler.SetUp).Methods("POST")
	// router.HandleFunc("/nemesis/{name}/teardown", nemesisHandler.TearDown).Methods("POST")
	router.HandleFunc("/nemesis/{name}/run", nemesisHandler.Run).Methods("POST")

	dbHandler := newDBHanlder(agent, rd)
	router.HandleFunc("/db/{name}/setup", dbHandler.SetUp).Methods("POST")
	router.HandleFunc("/db/{name}/teardown", dbHandler.TearDown).Methods("POST")

	return router
}

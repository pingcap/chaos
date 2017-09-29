package node

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/unrolled/render"
)

type nemesisHandler struct {
	agent *Agent
	rd    *render.Render
}

func newNemesisHandler(agent *Agent, rd *render.Render) *nemesisHandler {
	return &nemesisHandler{
		agent: agent,
		rd:    rd,
	}
}

func (h *nemesisHandler) getNemesis(w http.ResponseWriter, vars map[string]string) core.Nemesis {
	name := vars["name"]
	nemesis := core.GetNemesis(name)
	if nemesis == nil {
		h.rd.JSON(w, http.StatusNotFound, fmt.Sprintf("nemesis %s is not registered", name))
		return nil
	}
	return nemesis
}

func (h *nemesisHandler) Run(w http.ResponseWriter, r *http.Request) {
	h.agent.nemesisLock.Lock()
	defer h.agent.nemesisLock.Unlock()

	vars := mux.Vars(r)
	nemesis := h.getNemesis(w, vars)
	if nemesis == nil {
		return
	}

	node := r.FormValue("node")
	invokeArgs := strings.Split(r.FormValue("invoke_args"), ",")
	recoverArgs := strings.Split(r.FormValue("recover_args"), ",")
	runTime, _ := time.ParseDuration(r.FormValue("dur"))
	if runTime == 0 {
		runTime = time.Second * time.Duration(rand.Intn(10)+1)
	}

	log.Printf("invoke nemesis %s with %v on node %s", nemesis.Name(), invokeArgs, node)

	defer func() {
		log.Printf("recover nemesis %s with %v on node %s", nemesis.Name(), recoverArgs, node)
		nemesis.Recover(h.agent.ctx, node, recoverArgs...)
	}()

	if err := nemesis.Invoke(h.agent.ctx, node, invokeArgs...); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	select {
	case <-h.agent.ctx.Done():
	case <-time.After(runTime):
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

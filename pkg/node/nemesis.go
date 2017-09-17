package node

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/unrolled/render"
)

type nemesisHandler struct {
	n  *Node
	rd *render.Render
}

func newNemesisHandler(n *Node, rd *render.Render) *nemesisHandler {
	return &nemesisHandler{
		n:  n,
		rd: rd,
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
	h.n.nemesisLock.Lock()
	defer h.n.nemesisLock.Unlock()

	vars := mux.Vars(r)
	nemesis := h.getNemesis(w, vars)
	if nemesis == nil {
		return
	}

	node := r.FormValue("node")
	args := strings.Split(r.FormValue("args"), ",")
	runTime, _ := time.ParseDuration(r.FormValue("dur"))
	if runTime == 0 {
		runTime = 10 * time.Second
	}

	log.Printf("invoke nemesis %s with %v on node %s", nemesis.Name(), args, node)

	defer nemesis.Recover(h.n.ctx, node)

	if err := nemesis.Invoke(h.n.ctx, node, args...); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	select {
	case <-h.n.ctx.Done():
	case <-time.After(runTime):
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

package node

import (
	"strings"
	"fmt"
	"net/http"

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

func (h *nemesisHandler) SetUp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nemesis := h.getNemesis(w, vars)
	if nemesis == nil {
		return
	}

	if err := nemesis.SetUp(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *nemesisHandler) TearDown(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nemesis := h.getNemesis(w, vars)
	if nemesis == nil {
		return
	}

	if err := nemesis.TearDown(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *nemesisHandler) Invoke(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nemesis := h.getNemesis(w, vars)
	if nemesis == nil {
		return
	}

	args := strings.Split(r.FormValue("args"), ",")
	if err := nemesis.Invoke(h.n.ctx, h.n.name, args...); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

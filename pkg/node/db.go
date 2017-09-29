package node

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/unrolled/render"
)

type dbHandler struct {
	agent  *Agent
	rd *render.Render
}

func newDBHanlder(agent *Agent, rd *render.Render) *dbHandler {
	return &dbHandler{
		agent:  agent,
		rd: rd,
	}
}

func (h *dbHandler) getDB(w http.ResponseWriter, vars map[string]string) core.DB {
	name := vars["name"]
	db := core.GetDB(name)
	if db == nil {
		h.rd.JSON(w, http.StatusNotFound, fmt.Sprintf("db %s is not registered", name))
		return nil
	}
	return db
}

func (h *dbHandler) SetUp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}
	node := r.FormValue("node")
	nodes := strings.Split(r.FormValue("nodes"), ",")

	log.Printf("set up db %s on node %s", db.Name(), node)
	if err := db.SetUp(h.agent.ctx, nodes, node); err != nil {
		log.Panicf("set up db %s failed %v", db.Name(), err)
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *dbHandler) TearDown(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}

	node := r.FormValue("node")
	nodes := strings.Split(r.FormValue("nodes"), ",")

	log.Printf("tear down db %s on node %s", db.Name(), node)
	if err := db.TearDown(h.agent.ctx, nodes, node); err != nil {
		log.Panicf("tear down db %s failed %v", db.Name(), err)
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

package node

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/unrolled/render"
)

type dbHandler struct {
	n  *Node
	rd *render.Render
}

func newDBHanlder(n *Node, rd *render.Render) *dbHandler {
	return &dbHandler{
		n:  n,
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

	if err := db.SetUp(h.n.ctx, h.n.name); err != nil {
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

	if err := db.TearDown(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *dbHandler) Start(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}

	if err := db.Start(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *dbHandler) Stop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}

	if err := db.Stop(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *dbHandler) Kill(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}

	if err := db.Kill(h.n.ctx, h.n.name); err != nil {
		h.rd.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

func (h *dbHandler) IsRunning(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db := h.getDB(w, vars)
	if db == nil {
		return
	}

	if !db.IsRunning(h.n.ctx, h.n.name) {
		h.rd.JSON(w, http.StatusNotFound, fmt.Sprintf("db %s is not running", db.Name()))
		return
	}

	h.rd.JSON(w, http.StatusOK, nil)
}

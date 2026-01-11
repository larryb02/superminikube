package pod

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/utils"
)

func NewHandler(service Service) handler {
	return handler{
		service: service,
	}
}

type handler struct {
	service Service
}

func (h *handler) GetPod(w http.ResponseWriter, r *http.Request) {
	nodename := r.URL.Query().Get("nodename")
	if nodename == "" {
		http.Error(w, "nodename required", http.StatusBadRequest)
		return
	}
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "uid required", http.StatusBadRequest)
		return
	}
	// TODO: custom errors for potential cases (key not found or internal error)
	pod, err := h.service.GetPodByUid(r.Context(), nodename, uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pod)
}

func (h *handler) ListPods(w http.ResponseWriter, r *http.Request) {
	pods, err := h.service.ListAllNamespacePods(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pods)
}

func (h *handler) CreatePod(w http.ResponseWriter, r *http.Request) {
	nodename := r.URL.Query().Get("nodename")
	if nodename == "" {
		http.Error(w, "nodename required", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var spec api.PodSpec
	err := json.NewDecoder(r.Body).Decode(&spec)
	// TODO: better request body handling
	if err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "Empty request body", http.StatusBadRequest)
		} else {
			http.Error(w, "Malformed request", http.StatusBadRequest)
			slog.Error("failed to decode", "msg", err)
		}
		return
	}
	slog.Debug("request body", "body", spec)
	pod, err := h.service.CreatePod(r.Context(), nodename, spec)
	if err != nil {
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(pod)
	utils.WriteJSONResponse(w, http.StatusCreated, pod)
}

func (h *handler) DeletePod(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("DELETE Pod\n"))
}

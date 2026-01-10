package apiserver

import (
	"encoding/json"
	"net/http"

	"superminikube/pkg/api"
	"superminikube/pkg/apiserver/watch"
)

func (s *APIServer) PodHandler(w http.ResponseWriter, r *http.Request) {
	nodename := r.URL.Query().Get("nodename")
	if nodename == "" {
		http.Error(w, "nodename required", http.StatusBadRequest)
		return
	}
	// handle get and post methods
	switch r.Method {
	case http.MethodGet:
		// vars := mux.Vars(r)
		// TODO: do some param parsing here
		uid := r.URL.Query().Get("uid")
		if uid != "" {
			// TODO: custom errors for potential cases (key not found or internal error)
			pod, err := GetPodByUid(r.Context(), nodename, uid, s.redisClient)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(pod)
		} else {
			pods, err := ListAllNamespacePods(s.redisClient)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(pods)
		}
	case http.MethodPost:
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
		pod, err := CreatePod(
			r.Context(),
			nodename,
			spec,
			s.redisClient,
		)
		if err != nil {
			http.Error(w, "Failed to process request", http.StatusInternalServerError)
			return
		}
		// TODO: s.watchService.Notify only works because PodHandler is currently of type APIServer,
		// this is bad coupling in my opinion
		// also 'what if' this method gets called while no watcher assigned to endpoint what happens then?
		// SHOULDNT happen but should prepared
		// this should also just get called inside the CreatePod function doesn't belong out here
		err = s.watchService.Notify(watch.WatchEvent{
			EventType: watch.Add,
			Resource:  "pod",
			Node:      nodename,
			Pod:       pod,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pod)
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("DELETE Pod\n"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

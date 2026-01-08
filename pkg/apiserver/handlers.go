package apiserver

import (
	"encoding/json"
	"net/http"

	"superminikube/pkg/apiserver/watch"
	"superminikube/pkg/spec"
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
			pod, err := GetPodByUid(nodename, uid, s.redisClient)
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
		// TODO: s.watchService.Notify only works because PodHandler is currently of type APIServer,
		// this is bad coupling in my opinion
		// also 'what if' this method gets called while no watcher assigned to endpoint what happens then?
		// SHOULDNT happen but should prepared
		err := s.watchService.Notify(watch.WatchEvent{
			EventType:     watch.Add,
			Resource: "pod",
			Node:     nodename,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		pod, err := CreatePod(
			nodename,
			&spec.ContainerSpec{
				Image: "nginx",
			},
			s.redisClient,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

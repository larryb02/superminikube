package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (ws *WatchService) WatchHandler(w http.ResponseWriter, r *http.Request) {
	// want this to be a stream
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher, ok := w.(http.Flusher)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel() // Ensure that cancel is called when done
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	// couple cases
	// on connect (client hits /watch endpoint)
	// for now nothing special happens
	// dispatch (send received event to client side)
	// on disconnect
	// aka cancelled req context
	// on ctx.Done()
	// cleanly close client connections
	nodename := r.URL.Query().Get("nodename")
	// NOTE: in theory an empty value here shouldn't cause a problem
	if nodename == "" {
		http.Error(w, "nodename required", http.StatusBadRequest)
		return
	}
	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()
	ch := ws.Watch(fmt.Sprintf("pod/%s", nodename))
	for {
		select {
		case ev := <-ch:
			slog.Info(fmt.Sprintf("received event: %v", ev))
			b, err := json.Marshal(ev)
			if err != nil {
				slog.Error(fmt.Sprintf("error marshaling: %v", err))
				continue
			}
			_, err = fmt.Fprintf(w, "data: %v\n\n", string(b))
			if err != nil {
				slog.Error(fmt.Sprintf("error writing response: %v", err))
			}
			flusher.Flush()
		case <-keepAliveTicker.C:
			_, err := w.Write([]byte(":keepalive\n\n"))
			if err != nil {
				slog.Error(fmt.Sprintf("error writing response: %v", err))
			}
			flusher.Flush()
		case <-ctx.Done():
			slog.Debug("request context done")
			//TODO: close and delete channel
			return
		// case <-parentCtx.Done():
		// 	slog.Debug("watch service context done")
		// 	// TODO: send an event so client can disconnect cleanly
		// 	return
		// }
		}
	}
}

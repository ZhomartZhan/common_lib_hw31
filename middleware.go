package common_lib_hw31

import (
	"context"
	"encoding/json"
	"net/http"
)

type Middleware interface {
	LoginMiddleware(fn http.HandlerFunc) http.HandlerFunc
}

type middleware struct {
	redisConnectionStore RedisConnectStore
}

func NewMiddleware(r RedisConnectStore) Middleware {
	return &middleware{redisConnectionStore: r}
}

func (m *middleware) LoginMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			respondJSON(w, http.StatusBadRequest, &ErrorMessage{
				Message: "For access needed authorization header",
				Status:  http.StatusBadRequest,
			})
			return
		} else {
			userId, err := m.redisConnectionStore.Get(token)
			if err != nil && err.Error() == "redis: nil" {
				respondJSON(w, http.StatusBadRequest, &ErrorMessage{
					Message: "Your access key is expired",
					Status:  http.StatusBadRequest,
				})
				return
			} else if err != nil {
				respondJSON(w, http.StatusBadRequest, &ErrorMessage{
					Message: err.Error(),
					Status:  http.StatusBadRequest,
				})
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", userId)
			r = r.WithContext(ctx)
		}
		fn.ServeHTTP(w, r)
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

type ErrorMessage struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/conn"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/model"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/repo"
)

type createUserRequest struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	repo := repo.NewUserPostgresRepository(conn.DefaultDB())

	uID := chi.URLParam(r, "id")
	userID, _ := strconv.Atoi(uID)

	if user, err := repo.Get(userID); err == nil {
		if err := json.NewEncoder(w).Encode(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if user == nil {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("user not found"))
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	w.Write([]byte(fmt.Sprintf("failed to get user with id: %s", uID)))
	w.WriteHeader(http.StatusNoContent)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := &createUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user := &model.User{
		ID:    req.ID,
		Name:  req.Name,
		Title: req.Title,
	}

	repo := repo.NewUserPostgresRepository(conn.DefaultDB())
	if err := repo.Update(user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("update successful"))
}

package handlers

import (
	"faucet-svc/doorman"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func Authenticate(w http.ResponseWriter, r *http.Request) {
	_, ok := doorman.GetHeader(r, "Authorization")
	if !ok {
		w.WriteHeader(401)
		ape.Render(w, problems.Unauthorized())
		return
	}
	ape.Render(w, resources.User{Id: "1", Type: "user"})
}

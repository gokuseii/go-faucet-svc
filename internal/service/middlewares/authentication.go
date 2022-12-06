package middlewares

import (
	"faucet-svc/internal/service/handlers"
	"faucet-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func ValidateJwt(r *http.Request) (user *resources.User, err error) {
	doorman := handlers.DoormanConnector(r)
	user, err = doorman.Authenticate(r)
	return
}

func CheckAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := ValidateJwt(r); err != nil {
			ape.RenderErr(w, problems.Unauthorized())
			return
		}
		next.ServeHTTP(w, r)
	})
}

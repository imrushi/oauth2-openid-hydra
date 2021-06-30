package handler

import (
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "PostLogin")
	defer func() {
		span.Finish()
		ctx.Done()
	}()

	formData := struct {
		LoginChallenge string `validate:"required"`
		Email          string `validate:"required"`
		Password       string `validate:"required"`
		RememberMe     string `validate:"required"`
	}{
		LoginChallenge: r.FormValue("login_challenge"),
		Email:          r.FormValue("email"),
		Password:       r.FormValue("password"),
		RememberMe:     r.FormValue("remember_me"),
	}

	// TODO validation

	var rememberMe = formData.RememberMe == "ture"
	// // user, err := h.UserRepo.GetUserByEmail(r.Context(), formData.Email)
	// if err != nil {
	// 	http.Error(w, "User not found", http.StatusNotFound)
	// }

	// if user.Password != formData.Password {
	// 	http.Error(w, "Wrong username and password", http.StatusNotFound)
	// }

	ok, err := h.Um.Authenticate(r.Context(), formData.Email, formData.Password)
	if err != nil {
		http.Error(w, "User not found "+err.Error(), http.StatusNotFound)
	}
	if !ok {
		http.Error(w, "Invalid Credntials", http.StatusNotFound)
	}
	//using Hydara admin to accept login request!
	loginGetParam := admin.NewGetLoginRequestParams()
	loginGetParam.SetLoginChallenge(formData.LoginChallenge)

	_, err = h.HydraAdmin.GetLoginRequest(loginGetParam)
	if err != nil {
		//if error, redirects to ....
		str := fmt.Sprint("error GetLoginRequest", err.Error())
		http.Error(w, str, http.StatusUnprocessableEntity)
	}

	subject := fmt.Sprint(formData.Email)

	loginAcceptParam := admin.NewAcceptLoginRequestParams()
	loginAcceptParam.WithContext(ctx)
	loginAcceptParam.SetLoginChallenge(formData.LoginChallenge)
	loginAcceptParam.SetBody(&models.AcceptLoginRequest{
		Subject:  &subject,
		Remember: rememberMe,
	})

	respLoginAccept, err := h.HydraAdmin.AcceptLoginRequest(loginAcceptParam)
	if err != nil {
		//if error, redirects to....
		str := fmt.Sprint("error AcceptLoignRequest", err.Error())
		http.Error(w, str, http.StatusUnprocessableEntity)
	}

	//If success, it will redirect to consent page using handler GetConsent
	//It then show the consent form
	http.Redirect(w, r, *respLoginAccept.GetPayload().RedirectTo, http.StatusFound)
}

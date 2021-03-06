package handler

import (
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "GetLogin")
	defer func() {
		span.Finish()
		ctx.Done()
	}()

	loginChallenge := strings.TrimSpace(r.URL.Query().Get("login_challenge"))
	if loginChallenge == "" {
		renderTemplate(w, "login.html", map[string]interface{}{
			"ErrorTitle":   "Login Challenge Is Not Exist!",
			"ErrorContent": "Login Challenge Is Not Exist!",
		})
	}

	//using Hydra Admin to get the login challenge info
	loginGetParam := admin.NewGetLoginRequestParams()
	loginGetParam.WithContext(ctx)
	loginGetParam.SetLoginChallenge(loginChallenge)

	respLoginGet, err := h.HydraAdmin.GetLoginRequest(loginGetParam)
	if err != nil {
		renderTemplate(w, "login.html", map[string]interface{}{
			"ErrorTitle":   "Failed When Get Login Request Info",
			"ErrorContent": err.Error(),
		})
	}

	skip := false
	if respLoginGet.GetPayload().Skip != nil {
		skip = *respLoginGet.GetPayload().Skip
	}

	// If hydra was already able to authenticate the user, skip will be true and we do not need to re-authenticate
	// the user.
	if skip {
		//Using Hydra Admin to accept login request!
		loginAcceptParam := admin.NewAcceptLoginRequestParams()
		loginAcceptParam.WithContext(ctx)
		loginAcceptParam.SetLoginChallenge(loginChallenge)
		loginAcceptParam.SetBody(&models.AcceptLoginRequest{
			Subject: respLoginGet.GetPayload().Subject,
		})

		respLoginAccept, err := h.HydraAdmin.AcceptLoginRequest(loginAcceptParam)
		if err != nil {
			renderTemplate(w, "login.html", map[string]interface{}{
				"ErrorTitle":   "Cannot Accept Login Request",
				"ErrorContent": err.Error(),
			})
		}

		// If success, it will redirect to consent page using handler GetConsent
		// If then show the consent form
		http.Redirect(w, r, *respLoginAccept.GetPayload().RedirectTo, http.StatusFound)
	}

	renderTemplate(w, "login.html", map[string]interface{}{
		"LoginChallenge": loginChallenge,
	})
}

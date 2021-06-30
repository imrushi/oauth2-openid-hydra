package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h Handler) GetConsent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "GetConsent")
	defer func() {
		span.Finish()
		ctx.Done()
	}()

	consentChallenge := strings.TrimSpace(r.URL.Query().Get("consent_challenge"))
	if consentChallenge == "" {
		renderTemplate(w, "consent.html", map[string]interface{}{
			"ErrorTitle":   "Cannot Accept Consent Request",
			"ErrorContent": "Consent challenge is empty",
		})
	}

	consentGetParamas := admin.NewGetConsentRequestParams()
	consentGetParamas.WithContext(ctx)
	consentGetParamas.SetConsentChallenge(consentChallenge)

	consentGetResp, err := h.HydraAdmin.GetConsentRequest(consentGetParamas)
	if err != nil {
		renderTemplate(w, "consent.html", map[string]interface{}{
			"ErrorTitle":   "Cannot Accept Consent Request",
			"ErrorContent": err.Error(),
		})
	}

	// If a user has granted this application the requested scope, hydra will tell us to not show the UI.
	if consentGetResp.GetPayload().Skip {
		// You can apply logic here, for example grant another scope, or do whatever...
		// ...

		// Now it's time to grant the consent request.
		// You could also deny the request if something went terribly wrong
		consentAcceptBody := &models.AcceptConsentRequest{
			GrantAccessTokenAudience: consentGetResp.GetPayload().RequestedAccessTokenAudience,
			GrantScope:               consentGetResp.GetPayload().RequestedScope,
		}

		consentAcceptParams := admin.NewAcceptConsentRequestParams()
		consentAcceptParams.WithContext(ctx)
		consentAcceptParams.SetConsentChallenge(consentChallenge)
		consentAcceptParams.WithBody(consentAcceptBody)

		consentAcceptResp, err := h.HydraAdmin.AcceptConsentRequest(consentAcceptParams)
		if err != nil {
			str := fmt.Sprint("error AcceptConsentRequest", err.Error())
			http.Error(w, str, http.StatusUnprocessableEntity)
		}

		http.Redirect(w, r, *consentAcceptResp.GetPayload().RedirectTo, http.StatusOK)
	}

	consentMessage := fmt.Sprintf("Application %s wants access resources on your behalf and to:", consentGetResp.GetPayload().Client.ClientName)

	renderTemplate(w, "consent.html", map[string]interface{}{
		"ConsentChallenge": consentChallenge,
		"ConsentMessage":   consentMessage,
		"RequestedScopes":  consentGetResp.GetPayload().RequestedScope,
	})
}

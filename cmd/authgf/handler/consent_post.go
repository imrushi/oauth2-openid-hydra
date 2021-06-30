package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
)

func (h Handler) PostConsent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	span, ctx := opentracing.StartSpanFromContext(ctx, "PostConsent")
	defer func() {
		span.Finish()
		ctx.Done()
	}()

	formData := struct {
		ConsentChallenge string   `validate:"required"`
		GrantScope       []string `validate:"requried"`
	}{
		ConsentChallenge: r.FormValue("consent_challenge"),
		GrantScope:       r.Form["grant_scope"],
	}

	consentGetParams := admin.NewGetConsentRequestParams()
	consentGetParams.WithContext(ctx)
	consentGetParams.SetConsentChallenge(formData.ConsentChallenge)

	consentGetResp, err := h.HydraAdmin.GetConsentRequest(consentGetParams)
	if err != nil {
		//if error, redirects to ...
		str := fmt.Sprint("error GetConsentRequest", err.Error())
		http.Error(w, str, http.StatusUnprocessableEntity)
	}
	log.Print(consentGetResp.Payload.Subject)
	claims, err := h.Um.FindOIDCClaims(r.Context(), consentGetResp.Payload.Subject)
	if err != nil {
		str := fmt.Sprint("Failed to find user's OIDC claims ", err.Error())
		http.Error(w, str, http.StatusInternalServerError)
	}
	log.Print("Found user's OIDC claims: ", claims)
	log.Print(formData.GrantScope, " ", consentGetResp.Payload.RequestedScope)

	grantScopeEl := formData.GrantScope
	claimsMap := make(map[string]string)
	for i := 0; i < len(grantScopeEl); i += 1 {
		claimsMap[grantScopeEl[i]] = grantScopeEl[i]
	}

	for claim := range claims {
		var found bool
		if scope := claimsMap[claim]; scope != "" {
			for _, rscope := range formData.GrantScope {
				log.Print(rscope)
				if rscope == scope {
					found = true
					break
				}
			}
		}
		if !found {
			delete(claims, claim)
			log.Print("Deleted the OIDC claim because it's not in requested scopes", "claim", claim)
		}
	}

	// for _, val := range grantScopeEl {
	// 	log.Print(val)
	// 	if claims[val] == val {
	// 		str := fmt.Sprintf("%v", claims[val])
	// 		log.Print(str)
	// 		grantScopeEl = append(grantScopeEl, str)
	// 	}
	// }
	// log.Print(grantScopeEl)
	// If a user has granted this application the requested scope, hydra will tell us to not show the UI.

	// Now it's time to grant the consent request. You could also deny the request if something went terribly wrong
	consentAcceptBody := &models.AcceptConsentRequest{
		GrantAccessTokenAudience: consentGetResp.GetPayload().RequestedAccessTokenAudience,
		GrantScope:               grantScopeEl,
		Session: &models.ConsentRequestSession{
			IDToken: claims,
		},
	}

	consentAcceptParams := admin.NewAcceptConsentRequestParams()
	consentAcceptParams.WithContext(ctx)
	consentAcceptParams.SetConsentChallenge(formData.ConsentChallenge)
	consentAcceptParams.WithBody(consentAcceptBody)

	consentAcceptResp, err := h.HydraAdmin.AcceptConsentRequest(consentAcceptParams)
	if err != nil {
		str := fmt.Sprint("error AcceptConsentReuqest", err.Error())
		http.Error(w, str, http.StatusUnprocessableEntity)
	}

	http.Redirect(w, r, *consentAcceptResp.GetPayload().RedirectTo, http.StatusFound)
}

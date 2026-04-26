package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/httpserver"
)

type AuthHandler struct {
	authSvc   *services.AuthService
	fositeSvc *services.FositeService
}

func NewAuthHandler(authSvc *services.AuthService, fositeSvc *services.FositeService) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		fositeSvc: fositeSvc,
	}
}

// Login API - Authenticates user and sets SSO cookie
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req ports.LoginCmd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	session, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		// Differentiate errors if needed (401 vs 500)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set SSO Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "citual_sso",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set to true in prod
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.authSvc.RequestPasswordReset(r.Context(), req.Email); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.authSvc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.authSvc.RequestMagicLink(r.Context(), req.Identifier); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *AuthHandler) LoginWithMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	session, err := h.authSvc.LoginWithMagicLink(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set SSO Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "citual_sso",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to dashboard or return success
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Redirect(w, r, "http://localhost:3000/dashboard", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := httpserver.GetUserID(ctx)
	tenantID := httpserver.GetTenantID(ctx)

	// Convert strings to UUIDs
	// In a real middleware we might have these as UUIDs in context or panic if missing
	// assuming AuthGuard ensures they are present and valid if the route is protected
	if userID == "" || tenantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uid, _ := uuid.Parse(userID)
	tid, _ := uuid.Parse(tenantID)

	me, err := h.authSvc.GetCurrentUser(ctx, uid, tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("GetMe: Returning user: %+v\n", me)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(me)
}

// Authorize Endpoint (OAuth2)
func (h *AuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Initialize Fosite Context
	ar, err := h.fositeSvc.NewAuthorizeRequest(ctx, r)
	if err != nil {
		h.fositeSvc.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	// 2. Check SSO Session
	cookie, err := r.Cookie("citual_sso")
	if err != nil || cookie.Value == "" {
		// Log the error for debugging
		fmt.Printf("Authorize: Missing Cookie. Err: %v\n", err)

		// No session -> Redirect to Frontend Login Page
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	fmt.Printf("Authorize: Received Cookie: %s\n", cookie.Value)

	decodedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		fmt.Printf("Authorize: Cookie Decode Error: %v\n", err)
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	session, err := h.authSvc.GetSession(ctx, decodedValue)
	if err != nil {
		// Log the error
		fmt.Printf("Authorize: Invalid Session. Err: %v\n", err)

		// Invalid session -> Redirect to Login
		http.Redirect(w, r, "http://localhost:3000/login?return_to="+url.QueryEscape(r.RequestURI), http.StatusFound)
		return
	}

	// CRITICAL: Inject TenantID into Context BEFORE calling services that touch RLS-enabled tables
	tid := session.TenantID.String()
	fmt.Printf("Authorize: Injecting TenantID: %s\n", tid)
	ctx = httpserver.SetTenantID(ctx, tid)

	// Fetch full user details for token claims
	user, err := h.authSvc.GetCurrentUser(ctx, session.UserID, session.TenantID)
	if err != nil {
		fmt.Printf("Authorize: Failed to fetch user details: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Approve Request
	userData := &ports.SessionUserData{
		UserID:       user.ID.String(),
		TenantID:     user.TenantID.String(),
		IsSuperAdmin: user.IsSuperAdmin,
		AuthzVersion: 1, // Default or fetch if managed
		Roles:        user.Roles,
	}

	resp, err := h.fositeSvc.NewAuthorizeResponse(ctx, ar, userData)
	if err != nil {
		fmt.Printf("Authorize: NewAuthorizeResponse Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Authorize: Fosite Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Authorize: Fosite Hint: %v\n", fositeErr.HintField)
		}
		h.fositeSvc.WriteAuthorizeError(ctx, w, ar, err)
		return
	}

	// 4. Write Response (Redirect with Code)
	h.fositeSvc.WriteAuthorizeResponse(ctx, w, ar, resp)
}

// Token Endpoint (OAuth2)
func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Create Access Request
	ar, err := h.fositeSvc.NewAccessRequest(ctx, r)
	if err != nil {
		fmt.Printf("Token: NewAccessRequest Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Token: Fosite Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Token: Fosite Hint: %v\n", fositeErr.HintField)
			fmt.Printf("Token: Fosite Desc: %v\n", fositeErr.DescriptionField)
			if fositeErr.Cause() != nil {
				fmt.Printf("Token: Fosite Cause: %v\n", fositeErr.Cause())
			}
		} else {
			fmt.Printf("Token: Non-Fosite Error Type: %T\n", err)
		}
		h.fositeSvc.WriteAccessError(ctx, w, ar, err)
		return
	}

	// PATCH: Helper to ensure TenantID and SuperAdmin status are in context for the Store
	if sess, ok := ar.GetSession().(*services.CitualSession); ok {
		if tid, ok := sess.Claims.Extra["tid"].(string); ok && tid != "" {
			ctx = httpserver.SetTenantID(ctx, tid)
		}
		if sa, ok := sess.Claims.Extra["sa"].(bool); ok {
			ctx = context.WithValue(ctx, httpserver.IsSuperAdminKey, sa)
		}
		if av, ok := sess.Claims.Extra["av"].(float64); ok {
			ctx = context.WithValue(ctx, httpserver.AuthzVersionKey, int(av))
		} else if av, ok := sess.Claims.Extra["av"].(int); ok {
			ctx = context.WithValue(ctx, httpserver.AuthzVersionKey, av)
		}
	} else {
		fmt.Printf("Token: Warning - Session is not services.CitualSession, got type: %T\n", ar.GetSession())
	}

	// 2. Create Access Response
	resp, err := h.fositeSvc.NewAccessResponse(ctx, ar)
	if err != nil {
		fmt.Printf("Token: NewAccessResponse Error: %v\n", err)
		if fositeErr, ok := err.(*fosite.RFC6749Error); ok {
			fmt.Printf("Token: Fosite Response Debug: %v\n", fositeErr.DebugField)
			fmt.Printf("Token: Fosite Response Hint: %v\n", fositeErr.HintField)
			fmt.Printf("Token: Fosite Response Description: %v\n", fositeErr.DescriptionField)
			if fositeErr.Cause() != nil {
				fmt.Printf("Token: Fosite Response Cause: %v\n", fositeErr.Cause())
			}
		}
		h.fositeSvc.WriteAccessError(ctx, w, ar, err)
		return
	}

	// 3. Write Response (JSON Token)
	h.fositeSvc.WriteAccessResponse(ctx, w, ar, resp)
}

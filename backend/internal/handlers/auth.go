package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"botmaker-backend/internal/models"
)

// sessionClaims is the signed payload of a session token.
type sessionClaims struct {
	UserID int64 `json:"uid"`
	Exp    int64 `json:"exp"`
}

const (
	sessionTTL   = 30 * 24 * time.Hour
	oauthStateTTL = 10 * time.Minute
)

// issueToken signs a session token for the given user.
func (a *API) issueToken(userID int64) (string, error) {
	claims := sessionClaims{UserID: userID, Exp: time.Now().Add(sessionTTL).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + a.signToken(body), nil
}

func (a *API) signToken(body string) string {
	mac := hmac.New(sha256.New, []byte(a.cfg.SessionSecret))
	_, _ = mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// authenticate resolves the platform account from the bearer token.
func (a *API) authenticate(r *http.Request) (*models.User, error) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil, errors.New("missing bearer token")
	}
	token := strings.TrimPrefix(header, "Bearer ")
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, errors.New("malformed token")
	}
	sig := a.signToken(parts[0])
	if subtle.ConstantTimeCompare([]byte(sig), []byte(parts[1])) != 1 {
		return nil, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("malformed token payload")
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("malformed token claims")
	}
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return a.store.GetUser(claims.UserID)
}

// requireUser guards a route with a valid session.
func (a *API) requireUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	u, err := a.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Avtorizatsiya talab qilinadi")
		return nil, false
	}
	return u, true
}

// requireAdmin guards a route with an admin session.
func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return nil, false
	}
	if u.Role != models.RoleAdmin && u.TelegramID != a.cfg.SuperAdminID {
		writeError(w, http.StatusForbidden, "Faqat adminlar uchun")
		return nil, false
	}
	return u, true
}

// isAdmin reports whether the account holds admin privileges.
func (a *API) isAdmin(u *models.User) bool {
	return u != nil && (u.Role == models.RoleAdmin || u.TelegramID == a.cfg.SuperAdminID)
}

// newOAuthState generates a random state value for the Google flow.
func (a *API) newOAuthState() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Prune expired states when the map grows.
	if len(a.oauthStates) > 500 {
		for k, exp := range a.oauthStates {
			if time.Now().After(exp) {
				delete(a.oauthStates, k)
			}
		}
	}
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)
	a.oauthStates[state] = time.Now().Add(oauthStateTTL)
	return state
}

// consumeOAuthState validates and removes a state value.
func (a *API) consumeOAuthState(state string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	exp, ok := a.oauthStates[state]
	if !ok {
		return false
	}
	delete(a.oauthStates, state)
	return time.Now().Before(exp)
}

// randHex returns n random bytes hex-encoded (2n characters).
func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---- Google OAuth ----

type googleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Aud           string `json:"aud"`
	Iss           string `json:"iss"`
	Exp           int64  `json:"exp"`
}

// verifyGoogleIDToken validates an ID token against Google's public
// tokeninfo endpoint and returns the profile claims.
func verifyGoogleIDToken(idToken string) (*googleTokenInfo, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("google tokeninfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("google tokeninfo: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("google tokeninfo: %w", err)
	}
	if info.Sub == "" {
		return nil, errors.New("google tokeninfo: missing subject")
	}
	return &info, nil
}

func (a *API) googleUserFromInfo(info *googleTokenInfo) (*models.User, error) {
	return a.store.UpsertGoogleUser(&models.User{
		GoogleID:  info.Sub,
		Email:     info.Email,
		FullName:  info.Name,
		AvatarURL: info.Picture,
		Role:      models.RoleUser,
	})
}

func (a *API) respondWithSession(w http.ResponseWriter, u *models.User) {
	token, err := a.issueToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Sessiya yaratib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

// handleGoogleAuth either starts the OAuth flow or exchanges a Google
// Identity Services ID token (sent as {"id_token": "..."}).
func (a *API) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req)

	if strings.TrimSpace(req.IDToken) != "" {
		info, err := verifyGoogleIDToken(strings.TrimSpace(req.IDToken))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Google tokeni yaroqsiz: "+err.Error())
			return
		}
		u, err := a.googleUserFromInfo(info)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Foydalanuvchi yaratib bo'lmadi")
			return
		}
		a.respondWithSession(w, u)
		return
	}

	state := a.newOAuthState()
	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + url.Values{
		"client_id":     {a.cfg.GoogleOAuth.ClientID},
		"redirect_uri":  {a.cfg.GoogleOAuth.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}.Encode()
	writeJSON(w, http.StatusOK, map[string]string{"auth_url": authURL})
}

// handleGoogleCallback completes the OAuth code exchange and redirects
// the browser back to the dashboard with a session token.
func (a *API) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// The dashboard is served under /botmaker/ (Next.js basePath).
	dashboard := strings.TrimSuffix(a.cfg.WebhookBaseURL, "/") + "/botmaker/"
	if a.cfg.WebhookBaseURL == "" {
		dashboard = "/botmaker/"
	}
	redirect := func(values url.Values) {
		http.Redirect(w, r, dashboard+"?"+values.Encode(), http.StatusFound)
	}

	state := r.URL.Query().Get("state")
	if !a.consumeOAuthState(state) {
		redirect(url.Values{"error": {"OAuth sessiyasi tugagan, qayta urinib ko'ring"}})
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		redirect(url.Values{"error": {"Google avtorizatsiyasi bekor qilindi"}})
		return
	}

	tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {a.cfg.GoogleOAuth.ClientID},
		"client_secret": {a.cfg.GoogleOAuth.ClientSecret},
		"redirect_uri":  {a.cfg.GoogleOAuth.RedirectURL},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		log.Printf("[auth] google token exchange: %v", err)
		redirect(url.Values{"error": {"Google token almashinuvi muvaffaqiyatsiz"}})
		return
	}
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(tokenResp.Body, 4096))
		log.Printf("[auth] google token exchange status %d: %s", tokenResp.StatusCode, body)
		redirect(url.Values{"error": {"Google token almashinuvi muvaffaqiyatsiz"}})
		return
	}
	var tokenBody struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenBody); err != nil || tokenBody.IDToken == "" {
		redirect(url.Values{"error": {"Google token almashinuvi muvaffaqiyatsiz"}})
		return
	}

	info, err := verifyGoogleIDToken(tokenBody.IDToken)
	if err != nil {
		log.Printf("[auth] google token verify: %v", err)
		redirect(url.Values{"error": {"Google tokeni yaroqsiz"}})
		return
	}
	u, err := a.googleUserFromInfo(info)
	if err != nil {
		log.Printf("[auth] google user upsert: %v", err)
		redirect(url.Values{"error": {"Foydalanuvchi yaratib bo'lmadi"}})
		return
	}
	token, err := a.issueToken(u.ID)
	if err != nil {
		redirect(url.Values{"error": {"Sessiya yaratib bo'lmadi"}})
		return
	}
	redirect(url.Values{"token": {token}})
}

// ---- Telegram auth ----

type telegramAuthRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
	BotToken  string `json:"bot_token"` // optional; all bots are tried otherwise
}

// telegramDataCheckString builds the canonical data-check-string of the
// Telegram Login Widget payload.
func telegramDataCheckString(req telegramAuthRequest) string {
	fields := map[string]string{
		"id": strconv.FormatInt(req.ID, 10),
	}
	if req.FirstName != "" {
		fields["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		fields["last_name"] = req.LastName
	}
	if req.Username != "" {
		fields["username"] = req.Username
	}
	if req.PhotoURL != "" {
		fields["photo_url"] = req.PhotoURL
	}
	if req.AuthDate > 0 {
		fields["auth_date"] = strconv.FormatInt(req.AuthDate, 10)
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(fields[k])
	}
	return sb.String()
}

// validTelegramHash reports whether the payload hash matches the bot
// token, following the Telegram Login Widget algorithm.
func validTelegramHash(req telegramAuthRequest, botToken string) bool {
	if req.Hash == "" {
		return false
	}
	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(telegramDataCheckString(req)))
	got := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(got), []byte(req.Hash)) == 1
}

// matchBotToken validates the payload against the provided bot token,
// falling back to every registered bot on the platform.
func (a *API) matchBotToken(req telegramAuthRequest) (bool, error) {
	if req.BotToken != "" && validTelegramHash(req, req.BotToken) {
		return true, nil
	}
	bots, err := a.store.ListBots()
	if err != nil {
		return false, err
	}
	for _, b := range bots {
		if b.Token != "" && validTelegramHash(req, b.Token) {
			return true, nil
		}
	}
	return false, nil
}

// handleTelegramAuth logs in with a Telegram Login Widget payload.
func (a *API) handleTelegramAuth(w http.ResponseWriter, r *http.Request) {
	var req telegramAuthRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	if req.ID == 0 {
		writeError(w, http.StatusBadRequest, "telegram_id kiritilmagan")
		return
	}
	if time.Now().Unix()-req.AuthDate > 24*3600 {
		writeError(w, http.StatusUnauthorized, "Telegram so'rovi eskirgan")
		return
	}
	ok, err := a.matchBotToken(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Botlar ro'yxatini o'qib bo'lmadi")
		return
	}
	if !ok {
		writeError(w, http.StatusUnauthorized, "Telegram autentifikatsiyasi muvaffaqiyatsiz")
		return
	}

	role := models.RoleUser
	if req.ID == a.cfg.SuperAdminID {
		role = models.RoleAdmin
	}
	fullName := strings.TrimSpace(req.FirstName + " " + req.LastName)
	u, err := a.store.UpsertTelegramUser(&models.User{
		TelegramID: req.ID,
		FullName:   fullName,
		AvatarURL:  req.PhotoURL,
		Role:       role,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Foydalanuvchi yaratib bo'lmadi")
		return
	}
	a.respondWithSession(w, u)
}

// handleQuickLogin logs in by Telegram ID without proof of ownership.
// It is a development convenience for the dashboard.
func (a *API) handleQuickLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		FullName   string `json:"full_name"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Username   string `json:"username"`
		AvatarURL  string `json:"avatar_url"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	if req.TelegramID == 0 {
		writeError(w, http.StatusBadRequest, "telegram_id kiritilmagan")
		return
	}
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		fullName = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}
	role := models.RoleUser
	if req.TelegramID == a.cfg.SuperAdminID {
		role = models.RoleAdmin
	}
	u, err := a.store.UpsertTelegramUser(&models.User{
		TelegramID: req.TelegramID,
		FullName:   fullName,
		AvatarURL:  req.AvatarURL,
		Role:       role,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Foydalanuvchi yaratib bo'lmadi")
		return
	}
	a.respondWithSession(w, u)
}

// handleMe returns the currently authenticated account.
func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, u)
}

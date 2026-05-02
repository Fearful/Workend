package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	Pool   *pgxpool.Pool
	Secure bool // set HTTPS-only cookie
}

type signupReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type meResp struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	IsAdmin     bool      `json:"is_admin"`
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		http.Error(w, "email, password, display_name required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var id uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Email, hash, req.DisplayName).Scan(&id)
	if err != nil {
		// 23505 = unique_violation
		if strings.Contains(err.Error(), "users_email_lower_idx") {
			http.Error(w, "email already in use", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.issueSession(w, r, id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(meResp{ID: id, Email: req.Email, DisplayName: req.DisplayName})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var (
		id           uuid.UUID
		passwordHash string
		displayName  string
		isAdmin      bool
	)
	err := h.Pool.QueryRow(r.Context(), `
		SELECT id, password_hash, display_name, is_admin
		FROM users
		WHERE lower(email) = $1
	`, req.Email).Scan(&id, &passwordHash, &displayName, &isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	ok, err := VerifyPassword(req.Password, passwordHash)
	if err != nil || !ok {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.issueSession(w, r, id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meResp{ID: id, Email: req.Email, DisplayName: displayName, IsAdmin: isAdmin})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		_ = DeleteSession(r.Context(), h.Pool, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	uid := UserID(r.Context())
	var resp meResp
	err := h.Pool.QueryRow(r.Context(), `
		SELECT id, email, display_name, is_admin FROM users WHERE id = $1
	`, uid).Scan(&resp.ID, &resp.Email, &resp.DisplayName, &resp.IsAdmin)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) issueSession(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	raw, expiresAt, err := CreateSession(r.Context(), h.Pool, userID, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    raw,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

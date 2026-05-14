package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// SigningKey represents an Ed25519 signing key stored per workspace.
type SigningKey struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	PublicKey   string     `json:"public_key"`
	KeyHash     string     `json:"key_hash"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
}

// RunSignature represents a cryptographic signature over a run.
type RunSignature struct {
	ID         uuid.UUID      `json:"id"`
	RunID      uuid.UUID      `json:"run_id"`
	KeyID      uuid.UUID      `json:"key_id"`
	Signature  string         `json:"signature"`
	Digest     string         `json:"digest"`
	SignedAt   time.Time      `json:"signed_at"`
	Provenance map[string]any `json:"provenance"`
	Verified   *bool          `json:"verified,omitempty"`
}

// Provenance holds SLSA-style metadata for a signed run.
type Provenance struct {
	BuildType   string `json:"buildType"`
	Builder     string `json:"builder"`
	RunID       string `json:"run_id"`
	TaskName    string `json:"task_name"`
	ProjectName string `json:"project_name"`
	CommitSHA   string `json:"commit_sha"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at"`
	ExitCode    int    `json:"exit_code"`
	LogDigest   string `json:"log_digest"`
}

// Handlers groups the HTTP handler methods for the signing feature.
type Handlers struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

// generateKeyResp includes the one-time private key in the response.
type generateKeyResp struct {
	SigningKey
	PrivateKey string `json:"private_key"`
}

// GenerateKey creates an Ed25519 keypair, stores the public key, and returns
// the private key in a one-time response.
// POST /api/workspaces/{id}/signing-keys
func (h *Handlers) GenerateKey(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		h.Logger.Error("generate ed25519 key", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pubB64 := base64.StdEncoding.EncodeToString(pub)
	hash := sha256.Sum256(pub)
	keyHash := base64.StdEncoding.EncodeToString(hash[:])

	var sk SigningKey
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO signing_keys (workspace_id, public_key, key_hash, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, workspace_id, public_key, key_hash, created_by, created_at, revoked_at
	`, wsID, pubB64, keyHash, uid).Scan(
		&sk.ID, &sk.WorkspaceID, &sk.PublicKey, &sk.KeyHash,
		&sk.CreatedBy, &sk.CreatedAt, &sk.RevokedAt)
	if err != nil {
		h.Logger.Error("insert signing key", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := generateKeyResp{
		SigningKey:  sk,
		PrivateKey: base64.StdEncoding.EncodeToString(priv),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ListKeys returns all signing keys for a workspace.
// GET /api/workspaces/{id}/signing-keys
func (h *Handlers) ListKeys(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, workspace_id, public_key, key_hash, created_by, created_at, revoked_at
		FROM signing_keys
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, wsID)
	if err != nil {
		h.Logger.Error("list signing keys", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []SigningKey{}
	for rows.Next() {
		var sk SigningKey
		if err := rows.Scan(&sk.ID, &sk.WorkspaceID, &sk.PublicKey, &sk.KeyHash,
			&sk.CreatedBy, &sk.CreatedAt, &sk.RevokedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, sk)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// RevokeKey sets revoked_at on a signing key.
// POST /api/signing-keys/{id}/revoke
func (h *Handlers) RevokeKey(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT workspace_id FROM signing_keys WHERE id = $1
	`, keyID).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	tag, err := h.Pool.Exec(r.Context(), `
		UPDATE signing_keys SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL
	`, keyID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "key already revoked", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// signRunReq is the JSON body for SignRun.
type signRunReq struct {
	KeyID      uuid.UUID `json:"key_id"`
	PrivateKey string    `json:"private_key"`
}

// SignRun signs a completed run with an Ed25519 key and stores SLSA provenance.
// POST /api/runs/{id}/sign
func (h *Handlers) SignRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	var req signRunReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Load run metadata (including workspace membership check).
	var (
		status      string
		taskName    string
		projectName string
		commitSHA   *string
		startedAt   *time.Time
		finishedAt  *time.Time
		exitCode    *int
		logPath     string
		wsID        uuid.UUID
	)
	err = h.Pool.QueryRow(r.Context(), `
		SELECT r.status, t.name, p.name, r.commit_sha,
		       r.started_at, r.finished_at, r.exit_code, r.log_path, p.workspace_id
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE r.id = $1 AND m.user_id = $2
	`, runID, uid).Scan(&status, &taskName, &projectName, &commitSHA,
		&startedAt, &finishedAt, &exitCode, &logPath, &wsID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.Logger.Error("load run for signing", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if status != "succeeded" && status != "failed" {
		http.Error(w, "run must be in terminal state (succeeded or failed)", http.StatusBadRequest)
		return
	}

	// Validate key belongs to workspace and is not revoked.
	var storedPubB64 string
	var keyWSID uuid.UUID
	var revokedAt *time.Time
	err = h.Pool.QueryRow(r.Context(), `
		SELECT workspace_id, public_key, revoked_at FROM signing_keys WHERE id = $1
	`, req.KeyID).Scan(&keyWSID, &storedPubB64, &revokedAt)
	if err != nil {
		http.Error(w, "signing key not found", http.StatusBadRequest)
		return
	}
	if keyWSID != wsID {
		http.Error(w, "signing key does not belong to this workspace", http.StatusBadRequest)
		return
	}
	if revokedAt != nil {
		http.Error(w, "signing key has been revoked", http.StatusBadRequest)
		return
	}

	// Decode and validate the provided private key.
	privBytes, err := base64.StdEncoding.DecodeString(req.PrivateKey)
	if err != nil || len(privBytes) != ed25519.PrivateKeySize {
		http.Error(w, "invalid private key", http.StatusBadRequest)
		return
	}
	privKey := ed25519.PrivateKey(privBytes)

	// Derive public key from private and compare with stored.
	derivedPub := privKey.Public().(ed25519.PublicKey)
	derivedPubB64 := base64.StdEncoding.EncodeToString(derivedPub)
	if derivedPubB64 != storedPubB64 {
		http.Error(w, "private key does not match stored public key", http.StatusBadRequest)
		return
	}

	// Compute log file SHA256.
	logDigest := ""
	if logPath != "" {
		logData, err := os.ReadFile(logPath)
		if err == nil {
			h := sha256.Sum256(logData)
			logDigest = fmt.Sprintf("sha256:%x", h)
		}
	}

	// Build content-to-sign.
	safeStr := func(s *string) string {
		if s != nil {
			return *s
		}
		return ""
	}
	safeTime := func(t *time.Time) string {
		if t != nil {
			return t.UTC().Format(time.RFC3339)
		}
		return ""
	}
	safeInt := func(i *int) int {
		if i != nil {
			return *i
		}
		return 0
	}

	content := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%s",
		runID.String(), taskName, projectName, safeStr(commitSHA),
		safeTime(startedAt), safeTime(finishedAt), safeInt(exitCode), logDigest)

	digest := sha256.Sum256([]byte(content))
	digestHex := fmt.Sprintf("sha256:%x", digest)

	// Sign the digest.
	sig := ed25519.Sign(privKey, digest[:])
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	// Build SLSA provenance.
	prov := Provenance{
		BuildType:   "https://workend.dev/run/v1",
		Builder:     "workend-runner",
		RunID:       runID.String(),
		TaskName:    taskName,
		ProjectName: projectName,
		CommitSHA:   safeStr(commitSHA),
		StartedAt:   safeTime(startedAt),
		FinishedAt:  safeTime(finishedAt),
		ExitCode:    safeInt(exitCode),
		LogDigest:   logDigest,
	}
	provJSON, err := json.Marshal(prov)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var rs RunSignature
	var provRaw []byte
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO run_signatures (run_id, key_id, signature, digest, provenance)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, run_id, key_id, signature, digest, signed_at, provenance
	`, runID, req.KeyID, sigB64, digestHex, provJSON).Scan(
		&rs.ID, &rs.RunID, &rs.KeyID, &rs.Signature, &rs.Digest, &rs.SignedAt, &provRaw)
	if err != nil {
		h.Logger.Error("insert run signature", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = json.Unmarshal(provRaw, &rs.Provenance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rs)
}

// VerifyRun loads an existing signature and verifies it against the public key.
// GET /api/runs/{id}/signature
func (h *Handlers) VerifyRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	// Membership check: user must be a workspace member who can see this run.
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id
		FROM runs r
		JOIN projects p ON p.id = r.project_id
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE r.id = $1 AND m.user_id = $2
	`, runID, uid).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var rs RunSignature
	var provRaw []byte
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, run_id, key_id, signature, digest, signed_at, provenance
		FROM run_signatures WHERE run_id = $1
	`, runID).Scan(&rs.ID, &rs.RunID, &rs.KeyID, &rs.Signature, &rs.Digest, &rs.SignedAt, &provRaw)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "no signature for this run", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = json.Unmarshal(provRaw, &rs.Provenance)

	// Load public key and verify.
	var pubB64 string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT public_key FROM signing_keys WHERE id = $1
	`, rs.KeyID).Scan(&pubB64)
	if err != nil {
		h.Logger.Error("load public key for verify", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	verified := false
	pubBytes, err := base64.StdEncoding.DecodeString(pubB64)
	if err == nil && len(pubBytes) == ed25519.PublicKeySize {
		sigBytes, serr := base64.StdEncoding.DecodeString(rs.Signature)
		if serr == nil {
			// Extract raw SHA256 hash from "sha256:<hex>" digest string.
			digestBytes, derr := hexDigestToBytes(rs.Digest)
			if derr == nil {
				verified = ed25519.Verify(ed25519.PublicKey(pubBytes), digestBytes, sigBytes)
			}
		}
	}
	rs.Verified = &verified

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rs)
}

// GetProvenance returns the SLSA provenance JSON for a signed run.
// GET /api/runs/{id}/provenance
func (h *Handlers) GetProvenance(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	// Membership check.
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id
		FROM runs r
		JOIN projects p ON p.id = r.project_id
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE r.id = $1 AND m.user_id = $2
	`, runID, uid).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var provRaw []byte
	err = h.Pool.QueryRow(r.Context(), `
		SELECT provenance FROM run_signatures WHERE run_id = $1
	`, runID).Scan(&provRaw)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "run is not signed", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Return as in-toto/SLSA-compatible envelope.
	var prov Provenance
	_ = json.Unmarshal(provRaw, &prov)

	envelope := map[string]any{
		"_type":          "https://in-toto.io/Statement/v0.1",
		"predicateType":  "https://slsa.dev/provenance/v0.2",
		"subject": []map[string]any{
			{
				"name": fmt.Sprintf("run:%s", runID.String()),
				"digest": map[string]string{
					"sha256": extractHex(prov.LogDigest),
				},
			},
		},
		"predicate": prov,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(envelope)
}

// hexDigestToBytes parses a "sha256:<hex>" string into raw bytes.
func hexDigestToBytes(digest string) ([]byte, error) {
	hex := extractHex(digest)
	if hex == "" {
		return nil, fmt.Errorf("invalid digest format: %s", digest)
	}
	out := make([]byte, len(hex)/2)
	for i := 0; i < len(out); i++ {
		hi := unhex(hex[i*2])
		lo := unhex(hex[i*2+1])
		if hi < 0 || lo < 0 {
			return nil, fmt.Errorf("invalid hex char in digest")
		}
		out[i] = byte(hi<<4 | lo)
	}
	return out, nil
}

func extractHex(digest string) string {
	const prefix = "sha256:"
	if len(digest) > len(prefix) && digest[:len(prefix)] == prefix {
		return digest[len(prefix):]
	}
	return ""
}

func unhex(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c-'a') + 10
	case 'A' <= c && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

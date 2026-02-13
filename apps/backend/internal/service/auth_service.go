package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	config       *config.Config
	db           *ent.Client
	tokenService *TokenService
	httpClient   *http.Client
}

type AuthResult struct {
	User   *ent.UserProfile
	Tokens *TokenPair
}

// NaverTokenResponse represents the Naver OAuth token response.
type NaverTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// NaverUserInfo represents the Naver user profile response.
type NaverUserInfo struct {
	ResultCode string `json:"resultcode"`
	Message    string `json:"message"`
	Response   struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
	} `json:"response"`
}

func NewAuthService(cfg *config.Config, db *ent.Client, tokenService *TokenService) *AuthService {
	return &AuthService{
		config:       cfg,
		db:           db,
		tokenService: tokenService,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetNaverAuthURL generates the Naver OAuth authorization URL.
func (s *AuthService) GetNaverAuthURL() (authURL, state string) {
	state = generateRandomState()
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {s.config.NaverClientID},
		"redirect_uri":  {s.config.NaverCallbackURL},
		"state":         {state},
	}
	return "https://nid.naver.com/oauth2.0/authorize?" + params.Encode(), state
}

// HandleNaverCallback processes the Naver OAuth callback.
func (s *AuthService) HandleNaverCallback(ctx context.Context, code, state string) (*AuthResult, error) {
	naverToken, err := s.exchangeNaverCode(code, state)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	naverUser, err := s.fetchNaverUserInfo(naverToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("user info fetch failed: %w", err)
	}

	user, err := s.findOrCreateNaverUser(ctx, naverUser)
	if err != nil {
		return nil, fmt.Errorf("user upsert failed: %w", err)
	}

	// Update last_login_at
	s.db.UserProfile.UpdateOneID(user.ID).
		SetLastLoginAt(time.Now()).
		ExecX(ctx)

	tokens, err := s.tokenService.IssueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("token issue failed: %w", err)
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
}

// Signup creates a new user with email/password.
func (s *AuthService) Signup(ctx context.Context, email, password, nickname string) (*AuthResult, error) {
	exists, _ := s.db.UserProfile.Query().
		Where(userprofile.EmailEQ(email)).
		Exist(ctx)
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("password hash failed: %w", err)
	}

	builder := s.db.UserProfile.Create().
		SetEmail(email).
		SetPasswordHash(string(hash)).
		SetAuthProvider(userprofile.AuthProviderEmail).
		SetRole(userprofile.RoleUser)
	if nickname != "" {
		builder = builder.SetNickname(nickname)
	}

	user, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	tokens, err := s.tokenService.IssueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
}

// Login authenticates a user with email/password.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.db.UserProfile.Query().
		Where(userprofile.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	s.db.UserProfile.UpdateOneID(user.ID).
		SetLastLoginAt(time.Now()).
		ExecX(ctx)

	tokens, err := s.tokenService.IssueTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
}

// RefreshTokens issues new tokens using a valid refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.db.UserProfile.Get(ctx, userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.tokenService.IssueTokenPair(user.ID, user.Role)
}

// GetUserByID fetches a user by ID.
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*ent.UserProfile, error) {
	return s.db.UserProfile.Get(ctx, userID)
}

func (s *AuthService) exchangeNaverCode(code, state string) (*NaverTokenResponse, error) {
	params := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.config.NaverClientID},
		"client_secret": {s.config.NaverClientSecret},
		"code":          {code},
		"state":         {state},
	}

	resp, err := s.httpClient.PostForm("https://nid.naver.com/oauth2.0/token", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp NaverTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("naver oauth error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &tokenResp, nil
}

func (s *AuthService) fetchNaverUserInfo(accessToken string) (*NaverUserInfo, error) {
	req, err := http.NewRequest("GET", "https://openapi.naver.com/v1/nid/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo NaverUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	if userInfo.ResultCode != "00" {
		return nil, fmt.Errorf("naver user info error: %s", userInfo.Message)
	}

	return &userInfo, nil
}

func (s *AuthService) findOrCreateNaverUser(ctx context.Context, info *NaverUserInfo) (*ent.UserProfile, error) {
	user, err := s.db.UserProfile.Query().
		Where(userprofile.NaverIDEQ(info.Response.ID)).
		Only(ctx)
	if err == nil {
		return user, nil
	}

	if !ent.IsNotFound(err) {
		return nil, err
	}

	builder := s.db.UserProfile.Create().
		SetNaverID(info.Response.ID).
		SetAuthProvider(userprofile.AuthProviderNaver).
		SetEmailVerified(true).
		SetRole(userprofile.RoleUser)

	if info.Response.Email != "" {
		builder = builder.SetEmail(info.Response.Email)
	}
	if info.Response.Nickname != "" {
		builder = builder.SetNickname(info.Response.Nickname)
	} else if info.Response.Name != "" {
		builder = builder.SetNickname(info.Response.Name)
	}

	return builder.Save(ctx)
}

func generateRandomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

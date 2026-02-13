package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthService(t *testing.T) (*AuthService, *TokenService) {
	t.Helper()
	db := testutil.NewTestClient(t)
	cfg := testutil.NewTestConfig()
	ts := NewTokenService(testutil.TestJWTSecret, testutil.TestAccessTokenTTL, testutil.TestRefreshTokenTTL)
	as := NewAuthService(cfg, db, ts)
	return as, ts
}

// --- GetNaverAuthURL ---

func TestGetNaverAuthURL_ReturnsValidURL(t *testing.T) {
	as, _ := newTestAuthService(t)

	authURL, state := as.GetNaverAuthURL()

	assert.Contains(t, authURL, "https://nid.naver.com/oauth2.0/authorize")
	assert.Contains(t, authURL, "client_id="+as.config.NaverClientID)
	assert.Contains(t, authURL, "state="+state)
	assert.NotEmpty(t, state)
	assert.Len(t, state, 32) // 16 bytes hex-encoded
}

// --- Signup ---

func TestSignup_Success(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	result, err := as.Signup(ctx, "test@example.com", "password123", "tester")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotNil(t, result.User)
	assert.Equal(t, "test@example.com", *result.User.Email)
	assert.Equal(t, "tester", result.User.Nickname)
	assert.Equal(t, userprofile.AuthProviderEmail, result.User.AuthProvider)
	assert.Equal(t, userprofile.RoleUser, result.User.Role)
	assert.NotNil(t, result.User.PasswordHash)

	assert.NotEmpty(t, result.Tokens.AccessToken)
	assert.NotEmpty(t, result.Tokens.RefreshToken)
	assert.Equal(t, int32(3600), result.Tokens.ExpiresIn)
}

func TestSignup_WithoutNickname(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	result, err := as.Signup(ctx, "no-nick@example.com", "password123", "")
	require.NoError(t, err)

	assert.Empty(t, result.User.Nickname)
}

func TestSignup_DuplicateEmail(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.Signup(ctx, "dup@example.com", "password123", "")
	require.NoError(t, err)

	_, err = as.Signup(ctx, "dup@example.com", "password456", "")
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.Signup(ctx, "login@example.com", "password123", "user1")
	require.NoError(t, err)

	result, err := as.Login(ctx, "login@example.com", "password123")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "login@example.com", *result.User.Email)
	assert.NotEmpty(t, result.Tokens.AccessToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.Signup(ctx, "wrongpw@example.com", "password123", "")
	require.NoError(t, err)

	_, err = as.Login(ctx, "wrongpw@example.com", "wrongpassword")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_NonExistentUser(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.Login(ctx, "noone@example.com", "password123")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_OAuthUserWithNoPassword(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	// Create a Naver OAuth user (no password hash)
	as.db.UserProfile.Create().
		SetNaverID("naver-123").
		SetEmail("naver@example.com").
		SetAuthProvider(userprofile.AuthProviderNaver).
		SetRole(userprofile.RoleUser).
		SaveX(ctx)

	_, err := as.Login(ctx, "naver@example.com", "anypassword")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// --- RefreshTokens ---

func TestRefreshTokens_Success(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	signupResult, err := as.Signup(ctx, "refresh@example.com", "password123", "")
	require.NoError(t, err)

	newTokens, err := as.RefreshTokens(ctx, signupResult.Tokens.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, newTokens)

	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	// Refresh tokens always differ (they have unique JTI)
	assert.NotEqual(t, signupResult.Tokens.RefreshToken, newTokens.RefreshToken)
}

func TestRefreshTokens_InvalidToken(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.RefreshTokens(ctx, "invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// --- GetUserByID ---

func TestGetUserByID_Success(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	signupResult, err := as.Signup(ctx, "getuser@example.com", "password123", "myuser")
	require.NoError(t, err)

	user, err := as.GetUserByID(ctx, signupResult.User.ID)
	require.NoError(t, err)
	assert.Equal(t, "getuser@example.com", *user.Email)
	assert.Equal(t, "myuser", user.Nickname)
}

func TestGetUserByID_NotFound(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	_, err := as.GetUserByID(ctx, [16]byte{})
	assert.Error(t, err)
}

// --- generateRandomState ---

func TestGenerateRandomState_Unique(t *testing.T) {
	s1 := generateRandomState()
	s2 := generateRandomState()

	assert.NotEqual(t, s1, s2)
	assert.Len(t, s1, 32)
}

// --- findOrCreateNaverUser ---

func TestFindOrCreateNaverUser_CreatesNew(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	info := &NaverUserInfo{ResultCode: "00", Message: "success"}
	info.Response.ID = "naver-new-user"
	info.Response.Email = "new@naver.com"
	info.Response.Nickname = "NaverUser"

	user, err := as.findOrCreateNaverUser(ctx, info)
	require.NoError(t, err)

	assert.Equal(t, "naver-new-user", *user.NaverID)
	assert.Equal(t, "new@naver.com", *user.Email)
	assert.Equal(t, "NaverUser", user.Nickname)
	assert.Equal(t, userprofile.AuthProviderNaver, user.AuthProvider)
	assert.True(t, user.EmailVerified)
}

func TestFindOrCreateNaverUser_FindsExisting(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	info := &NaverUserInfo{ResultCode: "00", Message: "success"}
	info.Response.ID = "naver-existing"
	info.Response.Email = "existing@naver.com"

	user1, err := as.findOrCreateNaverUser(ctx, info)
	require.NoError(t, err)

	user2, err := as.findOrCreateNaverUser(ctx, info)
	require.NoError(t, err)

	assert.Equal(t, user1.ID, user2.ID)
}

func TestFindOrCreateNaverUser_UsesNameWhenNoNickname(t *testing.T) {
	as, _ := newTestAuthService(t)
	ctx := context.Background()

	info := &NaverUserInfo{ResultCode: "00", Message: "success"}
	info.Response.ID = "naver-name-only"
	info.Response.Name = "RealName"

	user, err := as.findOrCreateNaverUser(ctx, info)
	require.NoError(t, err)

	assert.Equal(t, "RealName", user.Nickname)
}

// --- Integration: Full auth flow ---

func TestFullAuthFlow(t *testing.T) {
	as, ts := newTestAuthService(t)
	ctx := context.Background()

	// 1. Signup
	signupResult, err := as.Signup(ctx, "flow@example.com", "testpass1", "flowuser")
	require.NoError(t, err)

	// 2. Validate access token
	claims, err := ts.ValidateAccessToken(signupResult.Tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, signupResult.User.ID.String(), claims.Subject)
	assert.Equal(t, "user", claims.Role)

	// 3. Login with same credentials
	loginResult, err := as.Login(ctx, "flow@example.com", "testpass1")
	require.NoError(t, err)
	assert.Equal(t, signupResult.User.ID, loginResult.User.ID)

	// 4. Refresh tokens
	newTokens, err := as.RefreshTokens(ctx, loginResult.Tokens.RefreshToken)
	require.NoError(t, err)

	// 5. Validate new access token
	newClaims, err := ts.ValidateAccessToken(newTokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, signupResult.User.ID.String(), newClaims.Subject)

	// 6. GetUserByID
	user, err := as.GetUserByID(ctx, signupResult.User.ID)
	require.NoError(t, err)
	assert.Equal(t, "flow@example.com", *user.Email)
}

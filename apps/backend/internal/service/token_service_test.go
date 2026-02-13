package service

import (
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-minimum-32-chars!!"

func newTestTokenService() *TokenService {
	return NewTokenService(testSecret, 1*time.Hour, 7*24*time.Hour)
}

func TestIssueTokenPair_Success(t *testing.T) {
	ts := newTestTokenService()
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleUser)
	require.NoError(t, err)
	require.NotNil(t, pair)

	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, int32(3600), pair.ExpiresIn)

	// Verify access token claims
	claims, err := ts.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "colight", claims.Issuer)
}

func TestValidateAccessToken_Valid(t *testing.T) {
	ts := newTestTokenService()
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleAdmin)
	require.NoError(t, err)

	claims, err := ts.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, "admin", claims.Role)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	// Create a service with 0 TTL to produce already-expired tokens
	ts := NewTokenService(testSecret, -1*time.Second, 7*24*time.Hour)
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleUser)
	require.NoError(t, err)

	_, err = ts.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	ts := newTestTokenService()
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleUser)
	require.NoError(t, err)

	// Create a different service with a different secret
	otherTS := NewTokenService("different-secret-key-32-chars!!!", 1*time.Hour, 7*24*time.Hour)

	_, err = otherTS.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
}

func TestValidateRefreshToken_Valid(t *testing.T) {
	ts := newTestTokenService()
	userID := uuid.New()

	pair, err := ts.IssueTokenPair(userID, userprofile.RoleUser)
	require.NoError(t, err)

	claims, err := ts.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, "colight", claims.Issuer)
	// Refresh token should have a JTI
	assert.NotEmpty(t, claims.ID)
}

func TestValidateAccessToken_WrongSigningMethod(t *testing.T) {
	// Create a token signed with a different method (none)
	claims := jwt.MapClaims{"sub": "test", "role": "user"}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	ts := newTestTokenService()
	_, err := ts.ValidateAccessToken(tokenString)
	assert.Error(t, err)
}

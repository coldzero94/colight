import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Auth;

// ─── Models ───

model SignupRequest {
  email: string;
  password: string;
  nickname?: string;
}

model LoginRequest {
  email: string;
  password: string;
}

model AuthTokens {
  @encodedName("application/json", "access_token")
  accessToken: string;

  @encodedName("application/json", "refresh_token")
  refreshToken: string;

  @encodedName("application/json", "expires_in")
  expiresIn: int32;
}

model AuthResponse {
  user: UserInfo;
  tokens: AuthTokens;
}

model UserInfo {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";

  @encodedName("application/json", "auth_provider")
  authProvider: "email" | "naver";

  @encodedName("application/json", "onboarding_completed")
  onboardingCompleted: boolean;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

// ─── Auth Interface ───

@route("/v1/auth")
@tag("Auth")
interface AuthAPI {
  /** Naver OAuth start — redirects to Naver login page */
  @get
  @route("/naver/login")
  naverLogin(): void;

  /** Naver OAuth callback — exchanges code for tokens */
  @get
  @route("/naver/callback")
  naverCallback(@query code: string, @query state: string): void;

  /** Email signup */
  @post
  @route("/signup")
  signup(@body body: SignupRequest): {
    @statusCode statusCode: 201;
    @body body: { data: AuthResponse };
  } | Colight.Common.ErrorResponse;

  /** Email login */
  @post
  @route("/login")
  login(@body body: LoginRequest): {
    @statusCode statusCode: 200;
    @body body: { data: AuthResponse };
  } | Colight.Common.ErrorResponse;

  /** Refresh JWT tokens */
  @post
  @route("/refresh")
  refresh(@body body: { @encodedName("application/json", "refresh_token") refreshToken: string }): {
    @statusCode statusCode: 200;
    @body body: { data: AuthTokens };
  } | Colight.Common.ErrorResponse;

  /** Get current user info */
  @get
  @route("/me")
  @useAuth(BearerAuth)
  me(): {
    @statusCode statusCode: 200;
    @body body: { data: UserInfo };
  } | Colight.Common.ErrorResponse;

  /** Logout */
  @post
  @route("/logout")
  @useAuth(BearerAuth)
  logout(): {
    @statusCode statusCode: 200;
    @body body: { success: boolean };
  };
}

import { describe, it, expect } from "vitest";
import { loginSchema, signupSchema } from "../auth";

describe("loginSchema", () => {
  it("accepts valid input", () => {
    const result = loginSchema.safeParse({
      email: "test@example.com",
      password: "password123",
    });
    expect(result.success).toBe(true);
  });

  it("rejects invalid email", () => {
    const result = loginSchema.safeParse({
      email: "not-an-email",
      password: "password123",
    });
    expect(result.success).toBe(false);
  });

  it("rejects short password", () => {
    const result = loginSchema.safeParse({
      email: "test@example.com",
      password: "short",
    });
    expect(result.success).toBe(false);
  });
});

describe("signupSchema", () => {
  it("accepts valid input", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "password1",
      passwordConfirm: "password1",
    });
    expect(result.success).toBe(true);
  });

  it("accepts valid input with nickname", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "password1",
      passwordConfirm: "password1",
      nickname: "tester",
    });
    expect(result.success).toBe(true);
  });

  it("rejects password without digits", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "passwordonly",
      passwordConfirm: "passwordonly",
    });
    expect(result.success).toBe(false);
  });

  it("rejects password without letters", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "12345678",
      passwordConfirm: "12345678",
    });
    expect(result.success).toBe(false);
  });

  it("rejects mismatched passwords", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "password1",
      passwordConfirm: "different1",
    });
    expect(result.success).toBe(false);
  });

  it("rejects nickname over 50 characters", () => {
    const result = signupSchema.safeParse({
      email: "test@example.com",
      password: "password1",
      passwordConfirm: "password1",
      nickname: "a".repeat(51),
    });
    expect(result.success).toBe(false);
  });
});

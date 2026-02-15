import { describe, it, expect } from "vitest";
import {
  type UserRole,
  ROLE_LEVEL,
  hasRole,
  isManager,
  isAdmin,
  isSuperAdmin,
} from "@/lib/auth-utils";

describe("ROLE_LEVEL", () => {
  it("defines correct numeric levels", () => {
    expect(ROLE_LEVEL.user).toBe(1);
    expect(ROLE_LEVEL.manager).toBe(2);
    expect(ROLE_LEVEL.admin).toBe(3);
    expect(ROLE_LEVEL.super_admin).toBe(4);
  });
});

describe("hasRole", () => {
  it("returns true for equal role", () => {
    expect(hasRole("admin", "admin")).toBe(true);
  });

  it("returns true for higher role", () => {
    expect(hasRole("super_admin", "admin")).toBe(true);
    expect(hasRole("admin", "manager")).toBe(true);
    expect(hasRole("manager", "user")).toBe(true);
  });

  it("returns false for lower role", () => {
    expect(hasRole("user", "admin")).toBe(false);
    expect(hasRole("manager", "admin")).toBe(false);
    expect(hasRole("admin", "super_admin")).toBe(false);
  });
});

describe("role helper functions", () => {
  const roles: UserRole[] = ["user", "manager", "admin", "super_admin"];

  it("isManager returns true for manager and above", () => {
    expect(isManager("user")).toBe(false);
    expect(isManager("manager")).toBe(true);
    expect(isManager("admin")).toBe(true);
    expect(isManager("super_admin")).toBe(true);
  });

  it("isAdmin returns true for admin and above", () => {
    expect(isAdmin("user")).toBe(false);
    expect(isAdmin("manager")).toBe(false);
    expect(isAdmin("admin")).toBe(true);
    expect(isAdmin("super_admin")).toBe(true);
  });

  it("isSuperAdmin returns true only for super_admin", () => {
    expect(isSuperAdmin("user")).toBe(false);
    expect(isSuperAdmin("manager")).toBe(false);
    expect(isSuperAdmin("admin")).toBe(false);
    expect(isSuperAdmin("super_admin")).toBe(true);
  });
});

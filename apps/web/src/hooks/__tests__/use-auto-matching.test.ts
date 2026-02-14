import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { useAutoMatching } from "../use-auto-matching";

const mockMatchExperiences = vi.fn();

vi.mock("@/lib/api/matching", () => ({
  matchExperiences: (...args: unknown[]) => mockMatchExperiences(...args),
}));

describe("useAutoMatching", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("triggers matching when no existing result and experiences exist", async () => {
    mockMatchExperiences.mockResolvedValue({
      company_name: "테스트회사",
      matches: [{ experience_id: "1", overall_fit: 80 }],
      total: 1,
    });

    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: false,
        experienceCount: 3,
        enabled: true,
      })
    );

    // Should start matching
    expect(result.current.isMatching).toBe(true);

    await waitFor(() => expect(result.current.isMatching).toBe(false));

    expect(mockMatchExperiences).toHaveBeenCalledWith("테스트회사");
    expect(result.current.matchingResult).not.toBeNull();
    expect(result.current.matchingResult?.total).toBe(1);
  });

  it("skips matching when no experiences", () => {
    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: false,
        experienceCount: 0,
        enabled: true,
      })
    );

    expect(result.current.isMatching).toBe(false);
    expect(mockMatchExperiences).not.toHaveBeenCalled();
  });

  it("skips matching when result already exists", () => {
    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: true,
        experienceCount: 5,
        enabled: true,
      })
    );

    expect(result.current.isMatching).toBe(false);
    expect(mockMatchExperiences).not.toHaveBeenCalled();
  });

  it("shows loading state during matching", async () => {
    let resolvePromise: (value: unknown) => void;
    const promise = new Promise((resolve) => {
      resolvePromise = resolve;
    });
    mockMatchExperiences.mockReturnValue(promise);

    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: false,
        experienceCount: 1,
        enabled: true,
      })
    );

    // Should be loading
    expect(result.current.isMatching).toBe(true);

    // Resolve
    await act(async () => {
      resolvePromise!({
        company_name: "테스트회사",
        matches: [],
        total: 0,
      });
    });

    await waitFor(() => expect(result.current.isMatching).toBe(false));
  });

  it("handles matching error", async () => {
    mockMatchExperiences.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: false,
        experienceCount: 2,
        enabled: true,
      })
    );

    await waitFor(() => expect(result.current.isMatching).toBe(false));

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.message).toBe("Network error");
  });

  it("skips matching when disabled", () => {
    const { result } = renderHook(() =>
      useAutoMatching({
        companyName: "테스트회사",
        hasMatching: false,
        experienceCount: 3,
        enabled: false,
      })
    );

    expect(result.current.isMatching).toBe(false);
    expect(mockMatchExperiences).not.toHaveBeenCalled();
  });
});

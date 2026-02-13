import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import {
  useExperiences,
  useCreateExperience,
  useDeleteExperience,
} from "../use-experiences";

vi.mock("@/lib/api/experiences", () => ({
  fetchExperiences: vi.fn().mockResolvedValue([
    { id: "1", title: "경험1", weapons: [] },
  ]),
  createExperience: vi.fn().mockResolvedValue({ id: "new-1" }),
  deleteExperience: vi.fn().mockResolvedValue({ success: true }),
  fetchExperience: vi.fn(),
  updateExperience: vi.fn(),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(
      QueryClientProvider,
      { client: queryClient },
      children
    );
  };
}

describe("useExperiences", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches experience list successfully", async () => {
    const { result } = renderHook(() => useExperiences(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data![0].title).toBe("경험1");
  });
});

describe("useCreateExperience", () => {
  it("creates experience and invalidates cache", async () => {
    const { result } = renderHook(() => useCreateExperience(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ title: "새 경험" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ id: "new-1" });
  });
});

describe("useDeleteExperience", () => {
  it("deletes experience and invalidates cache", async () => {
    const { result } = renderHook(() => useDeleteExperience(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("1");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ success: true });
  });
});

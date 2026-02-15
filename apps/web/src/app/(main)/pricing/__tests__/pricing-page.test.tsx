import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import PricingPage from "../page";

vi.mock("sonner", () => ({
  toast: { info: vi.fn() },
}));

vi.mock("@/hooks/use-usage", () => ({
  useUsage: () => ({ data: { plan: "free", features: {} } }),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

describe("PricingPage", () => {
  it("renders all plan cards", () => {
    render(<PricingPage />, { wrapper: createWrapper() });
    expect(screen.getByText("Free")).toBeInTheDocument();
    expect(screen.getByText("Starter")).toBeInTheDocument();
    expect(screen.getByText("Pro")).toBeInTheDocument();
    expect(screen.getByText("Season")).toBeInTheDocument();
  });

  it("marks free plan as current", () => {
    render(<PricingPage />, { wrapper: createWrapper() });
    expect(screen.getByText("현재 플랜")).toBeDisabled();
  });

  it("shows coming soon toast for paid plans", async () => {
    const user = userEvent.setup();
    render(<PricingPage />, { wrapper: createWrapper() });

    const startButtons = screen.getAllByText("시작하기");
    await user.click(startButtons[0]);
    expect(toast.info).toHaveBeenCalledWith("결제 기능은 곧 출시됩니다!");
  });
});

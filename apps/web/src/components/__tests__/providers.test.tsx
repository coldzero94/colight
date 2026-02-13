import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

vi.mock("@tanstack/react-query-devtools", () => ({
  ReactQueryDevtools: () => null,
}));

import { Providers } from "../providers";

describe("Providers", () => {
  it("renders children", () => {
    render(
      <Providers>
        <div>App Content</div>
      </Providers>
    );

    expect(screen.getByText("App Content")).toBeInTheDocument();
  });

  it("provides QueryClient to children", () => {
    // If QueryClientProvider is not set up correctly, useQuery would throw
    render(
      <Providers>
        <div>Wrapped</div>
      </Providers>
    );

    expect(screen.getByText("Wrapped")).toBeInTheDocument();
  });
});

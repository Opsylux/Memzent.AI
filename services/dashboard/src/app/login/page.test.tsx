import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import LoginPage from "./page";

vi.mock("@/lib/supabase", () => ({
  supabase: {
    auth: {
      setSession: vi.fn().mockResolvedValue({}),
      onAuthStateChange: vi.fn(() => ({
        data: { subscription: { unsubscribe: vi.fn() } },
      })),
      signInWithOtp: vi.fn(),
      signInWithOAuth: vi.fn(),
    },
  },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    refresh: vi.fn(),
  }),
}));

describe("LoginPage", () => {
  it("renders welcome copy and email sign-in", () => {
    render(<LoginPage />);

    expect(screen.getByText("Sign in")).toBeInTheDocument();
    expect(
      screen.getByText("Access your org dashboard, keys, and playground")
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /GitHub/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Google/i })).toBeInTheDocument();
  });
});

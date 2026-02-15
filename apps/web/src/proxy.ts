import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function proxy(_request: NextRequest) {
  // JWT tokens are stored in localStorage (Zustand persist).
  // Server-side proxy cannot access localStorage, so route protection
  // is handled by the client-side AuthGuard component.
  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};

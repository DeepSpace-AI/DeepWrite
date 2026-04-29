"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { Sidebar } from "@/components/layout/sidebar";
import { Navbar } from "@/components/layout/navbar";
import { cn } from "@/lib/utils";

const publicRoutes = ["/login", "/register"];

export function MainLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, isHydrated } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const isPublicRoute = publicRoutes.includes(pathname || "");

  useEffect(() => {
    if (isHydrated && !isLoading && !isAuthenticated && !isPublicRoute) {
      router.push("/login");
    }
  }, [isHydrated, isLoading, isAuthenticated, isPublicRoute, router]);

  if (!isHydrated || isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen" style={{ background: "var(--bg-primary)" }}>
        <div
          className="w-8 h-8 border-4 border-t-transparent rounded-full animate-spin"
          style={{ borderColor: "var(--accent)", borderTopColor: "transparent" }}
        ></div>
      </div>
    );
  }

  if (isPublicRoute) {
    return <>{children}</>;
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen" style={{ background: "var(--bg-secondary)" }}>
      <Sidebar />
      <div className={cn("ml-64 transition-all duration-300")}>
        <Navbar />
        <main className="p-6">{children}</main>
      </div>
    </div>
  );
}

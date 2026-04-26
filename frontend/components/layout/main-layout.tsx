"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { Sidebar } from "@/components/layout/sidebar";
import { Navbar } from "@/components/layout/navbar";
import { cn } from "@/lib/utils";

const publicRoutes = ["/login", "/register"];

export function MainLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const isPublicRoute = publicRoutes.includes(pathname || "");

  useEffect(() => {
    if (!isLoading && !isAuthenticated && !isPublicRoute) {
      router.push("/login");
    }
  }, [isLoading, isAuthenticated, isPublicRoute, router]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
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
    <div className="min-h-screen bg-gray-50">
      <Sidebar />
      <div className={cn("ml-64 transition-all duration-300")}>
        <Navbar />
        <main className="p-6">{children}</main>
      </div>
    </div>
  );
}

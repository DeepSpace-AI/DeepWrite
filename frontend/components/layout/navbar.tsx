"use client";

import Link from "next/link";
import { useAuthStore } from "@/stores/auth";
import { Bell, Search, User, LogOut } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";

export function Navbar() {
  const { user, logout } = useAuthStore();

  return (
    <header
      className="sticky top-0 z-30 h-16"
      style={{ background: "var(--bg-card)", borderBottom: "1px solid var(--border)" }}
    >
      <div className="flex items-center justify-between h-full px-6">
        <div className="flex items-center flex-1 max-w-xl">
          <div className="relative w-full">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: "var(--text-muted)" }} />
            <input
              type="text"
              placeholder="搜索项目、文献..."
              className="w-full pl-10 pr-4 py-2 rounded-lg focus:outline-none focus:ring-2"
              style={{
                background: "var(--bg-secondary)",
                border: "1px solid var(--border)",
                color: "var(--text-primary)",
              }}
            />
          </div>
        </div>

        <div className="flex items-center gap-4">
          <ThemeToggle />

          <button
            className="relative p-2 rounded-lg"
            style={{ color: "var(--text-secondary)" }}
          >
            <Bell className="w-5 h-5" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
          </button>

          {user ? (
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <div
                  className="w-8 h-8 rounded-full flex items-center justify-center"
                  style={{ background: "var(--accent-light)", color: "var(--accent)" }}
                >
                  <User className="w-4 h-4" />
                </div>
                <span className="text-sm font-medium" style={{ color: "var(--text-primary)" }}>
                  {user.name}
                </span>
              </div>
              <button
                onClick={logout}
                className="p-2 rounded-lg"
                style={{ color: "var(--text-secondary)" }}
              >
                <LogOut className="w-4 h-4" />
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <Link
                href="/login"
                className="px-4 py-2 text-sm font-medium"
                style={{ color: "var(--text-secondary)" }}
              >
                登录
              </Link>
              <Link
                href="/register"
                className="px-4 py-2 text-sm font-medium text-white rounded-lg"
                style={{ background: "var(--accent)" }}
              >
                注册
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

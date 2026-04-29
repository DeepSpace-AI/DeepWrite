"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  FolderOpen,
  BookOpen,
  FileText,
  Code,
  Image,
  Send,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";

const mainNavItems = [
  { icon: LayoutDashboard, label: "工作台", href: "/dashboard" },
  { icon: FolderOpen, label: "项目", href: "/projects" },
];

const projectNavItems = [
  { icon: BookOpen, label: "文献", href: "references" },
  { icon: FileText, label: "写作", href: "writing" },
  { icon: Code, label: "代码", href: "code" },
  { icon: Image, label: "图像", href: "images" },
  { icon: Send, label: "投递", href: "submission" },
];

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const pathname = usePathname();

  const isProjectPage = pathname?.startsWith("/projects/") && pathname.split("/").length > 3;
  const projectId = isProjectPage ? pathname?.split("/")[2] : null;

  return (
    <aside
      className={cn(
        "fixed left-0 top-0 z-40 h-screen transition-all duration-300 flex flex-col",
        collapsed ? "w-16" : "w-64"
      )}
      style={{
        background: "var(--bg-sidebar)",
        borderRight: "1px solid var(--border)",
      }}
    >
      <div
        className="flex items-center justify-between h-16 px-4"
        style={{ borderBottom: "1px solid var(--border)" }}
      >
        {!collapsed && (
          <Link href="/dashboard" className="text-xl font-bold" style={{ color: "var(--text-primary)" }}>
            DeepWrite
          </Link>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="p-1 rounded-lg ml-auto"
          style={{ color: "var(--text-secondary)" }}
        >
          {collapsed ? (
            <ChevronRight className="w-5 h-5" />
          ) : (
            <ChevronLeft className="w-5 h-5" />
          )}
        </button>
      </div>

      <nav className="flex-1 py-4 space-y-1">
        {mainNavItems.map((item) => {
          const isActive = pathname === item.href || pathname?.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center px-4 py-3 mx-2 rounded-lg transition-colors"
              )}
              style={{
                background: isActive ? "var(--accent-light)" : "transparent",
                color: isActive ? "var(--accent)" : "var(--text-secondary)",
              }}
              title={collapsed ? item.label : undefined}
            >
              <item.icon className={cn("w-5 h-5", collapsed && "mx-auto")} />
              {!collapsed && (
                <span className="ml-3 font-medium">{item.label}</span>
              )}
            </Link>
          );
        })}

        {isProjectPage && projectId && (
          <>
            <div
              className={cn("mx-4 my-2", collapsed && "mx-2")}
              style={{ borderTop: "1px solid var(--border)" }}
            />
            {!collapsed && (
              <div className="px-4 py-2">
                <span
                  className="text-xs font-semibold uppercase tracking-wider"
                  style={{ color: "var(--text-muted)" }}
                >
                  项目功能
                </span>
              </div>
            )}
            {projectNavItems.map((item) => {
              const href = `/projects/${projectId}/${item.href}`;
              const isActive = pathname === href || pathname?.startsWith(href + "/");
              return (
                <Link
                  key={item.href}
                  href={href}
                  className={cn(
                    "flex items-center px-4 py-3 mx-2 rounded-lg transition-colors"
                  )}
                  style={{
                    background: isActive ? "var(--accent-light)" : "transparent",
                    color: isActive ? "var(--accent)" : "var(--text-secondary)",
                  }}
                  title={collapsed ? item.label : undefined}
                >
                  <item.icon className={cn("w-5 h-5", collapsed && "mx-auto")} />
                  {!collapsed && (
                    <span className="ml-3 font-medium">{item.label}</span>
                  )}
                </Link>
              );
            })}
          </>
        )}
      </nav>
    </aside>
  );
}

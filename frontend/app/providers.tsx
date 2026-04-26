"use client";

import { MainLayout } from "@/components/layout/main-layout";

export function Providers({ children }: { children: React.ReactNode }) {
  return <MainLayout>{children}</MainLayout>;
}

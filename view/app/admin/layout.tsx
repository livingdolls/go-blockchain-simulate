"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAdminStore } from "@/store/admin-store";
import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import { Toaster } from "@/components/ui/sonner";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, isHydrated } = useAdminStore();

  useEffect(() => {
    // Wait for store to hydrate from localStorage
    if (!isHydrated) {
      return;
    }

    // Allow login page to show even without auth
    if (pathname === "/admin/login") {
      return;
    }

    // Redirect to login if not authenticated (only after hydration)
    if (!isAuthenticated) {
      router.push("/admin/login");
    }
  }, [isAuthenticated, isHydrated, pathname, router]);

  // For login page, render without layout
  if (pathname === "/admin/login") {
    return (
      <>
        {children}
        <Toaster />
      </>
    );
  }

  // Wait for hydration before rendering protected pages
  if (!isHydrated) {
    return null;
  }

  // For other pages, check if authenticated before showing layout
  if (!isAuthenticated) {
    return null;
  }

  return (
    <SidebarProvider>
      <div className="flex h-screen overflow-hidden bg-background">
        <AppSidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <SiteHeader />
          <main className="flex-1 overflow-auto">
            <div className="p-6">{children}</div>
          </main>
        </div>
      </div>
      <Toaster />
    </SidebarProvider>
  );
}

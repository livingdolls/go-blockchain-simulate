"use client";

import { useAuthStore } from "@/store/auth-store";
import { useEffect } from "react";

export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const fetchUser = useAuthStore((state) => state.fetchUser);
  const loading = useAuthStore((state) => state.loading);

  useEffect(() => {
    fetchUser();
  }, []);

  if (loading) {
    return <div>Loading...</div>;
  }

  return <>{children}</>;
};

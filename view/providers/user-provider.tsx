"use client";

import { useAuthStore } from "@/store/auth-store";
import { useEffect } from "react";

export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const fetchUser = useAuthStore((state) => state.fetchUser);

  useEffect(() => {
    fetchUser();
  }, []);

  return <>{children}</>;
};

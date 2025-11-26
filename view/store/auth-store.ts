import { api } from "@/lib/axios";
import { TUser } from "@/types/user";
import { redirect } from "next/navigation";
import { create } from "zustand";

type AuthState = {
  user: TUser | null;
  fetchUser: () => Promise<void>;
  logout: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  fetchUser: async () => {
    try {
      const res = await api.get<TUser>("/profile");
      set({
        user: {
          address: res.data.address,
          public_key: res.data.public_key,
          balance: res.data.balance,
          username: res.data.username,
          id: res.data.id,
        },
      });
    } catch (error) {
      console.error("Failed to fetch user:", error);
      set({ user: null });
      redirect("/login");
    }
  },
  logout: () => {
    set({ user: null });
    redirect("/login");
  },
}));

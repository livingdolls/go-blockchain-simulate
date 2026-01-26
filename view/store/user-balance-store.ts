import { TUserBalance } from "@/types/balance";
import { create } from "zustand";

type UserBalanceStore = {
  userBalance: TUserBalance | null;
  setUserBalance: (balance: TUserBalance) => void;
  clearUserBalance: () => void;
};

export const useUserBalanceStore = create<UserBalanceStore>((set) => ({
  userBalance: null,

  setUserBalance: (balance) =>
    set(() => ({
      userBalance: balance,
    })),

  clearUserBalance: () =>
    set(() => ({
      userBalance: null,
    })),
}));

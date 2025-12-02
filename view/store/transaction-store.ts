import { TTransactionFilter } from "@/types/transaction";
import { create } from "zustand";

type TransactionStore = {
  // State
  filter: TTransactionFilter;

  // Actions
  setFilter: (filter: Partial<TTransactionFilter>) => void;
  updateFilter: (updates: Partial<TTransactionFilter>) => void;
  goToPage: (page: number) => void;
  changeLimit: (limit: number) => void;
  resetFilters: () => void;
};

const DEFAULT_FILTER: TTransactionFilter = {
  type: "all",
  status: "ALL",
  page: 1,
  limit: 10,
  sort_by: "id",
  order: "desc",
};

export const useTransactionStore = create<TransactionStore>((set) => ({
  // Initial State
  filter: DEFAULT_FILTER,

  // Actions
  setFilter: (newFilter) =>
    set(() => ({
      filter: { ...DEFAULT_FILTER, ...newFilter },
    })),

  updateFilter: (updates) =>
    set((state) => ({
      filter: { ...state.filter, ...updates },
    })),

  goToPage: (page) =>
    set((state) => ({
      filter: { ...state.filter, page },
    })),

  changeLimit: (limit) =>
    set((state) => ({
      filter: { ...state.filter, limit, page: 1 }, // Reset to page 1
    })),

  resetFilters: () =>
    set(() => ({
      filter: DEFAULT_FILTER,
    })),
}));

import { useTransactionStore } from "@/store/transaction-store";

export const useTransactionFilter = () => {
  const filter = useTransactionStore((state) => state.filter);
  const setFilter = useTransactionStore((state) => state.setFilter);
  const updateFilter = useTransactionStore((state) => state.updateFilter);
  const resetFilters = useTransactionStore((state) => state.resetFilters);
  const goToPage = useTransactionStore((state) => state.goToPage);
  const changeLimit = useTransactionStore((state) => state.changeLimit);

  return {
    filter,
    setFilter,
    updateFilter,
    resetFilters,
    goToPage,
    changeLimit,
  };
};

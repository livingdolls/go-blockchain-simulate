import { useTransactionStore } from "@/store/transaction-store";
import { TTransactionFilter } from "@/types/transaction";

export const useTransactionFullFilter = () => {
  const store = useTransactionStore();

  const defaultFilter: TTransactionFilter = {
    page: 1,
    limit: 10,
    type: "all",
    status: "ALL",
    sort_by: "created_at",
    order: "desc",
  };

  return {
    filter: store.filter,
    setFilter: store.setFilter,
    updateFilter: store.updateFilter,
    resetFilters: store.resetFilters,
    goToPage: store.goToPage,
    changeLimit: store.changeLimit,
    defaultFilter,
  };
};

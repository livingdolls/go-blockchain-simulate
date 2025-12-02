import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useTransactionStore } from "@/store/transaction-store";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

export const useTransaction = () => {
  const user = useAuthStore((state) => state.user);
  const filter = useTransactionStore((state) => state.filter);
  const setFilter = useTransactionStore((state) => state.setFilter);
  const updateFilter = useTransactionStore((state) => state.updateFilter);
  const resetFilters = useTransactionStore((state) => state.resetFilters);
  const goToPage = useTransactionStore((state) => state.goToPage);
  const changeLimit = useTransactionStore((state) => state.changeLimit);

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["transactions", filter],
    queryFn: async () => {
      if (!user) {
        toast.error("User not authenticated");
        return;
      }

      return await TransactionRepository.getTransactionByAddress(
        user?.address || "",
        filter.page,
        filter.limit,
        filter.type,
        filter.status,
        filter.order,
        filter.sort_by
      );
    },
    placeholderData: keepPreviousData,
  });

  return {
    transactions: data,
    isLoading,
    isError,
    refetch,
    filter,
    setFilter,
    updateFilter,
    resetFilters,
    goToPage,
    changeLimit,
    isFetching,
  };
};

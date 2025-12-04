import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useTransactionStore } from "@/store/transaction-store";
import { TTransactionWalletResponse } from "@/types/transaction";
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
        const emptyResponse: TTransactionWalletResponse = {
          balance: 0,
          address: "",
          transactions: {
            transactions: [],
            total: 0,
            page: filter.page,
            limit: filter.limit,
            total_pages: 1,
          },
        };
        return emptyResponse;
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

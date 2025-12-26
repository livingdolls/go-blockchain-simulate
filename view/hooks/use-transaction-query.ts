import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useTransactionStore } from "@/store/transaction-store";
import { keepPreviousData, useQuery } from "@tanstack/react-query";

export const useTransactionQuery = () => {
  const user = useAuthStore((s) => s.user);
  const filter = useTransactionStore((s) => s.filter);

  return useQuery({
    queryKey: ["transactions", filter],
    queryFn: async () => {
      return await TransactionRepository.getTransactionByAddress(
        user!.address,
        filter.page,
        filter.limit,
        filter.type,
        filter.status,
        filter.order,
        filter.sort_by
      );
    },
    placeholderData: keepPreviousData,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
};

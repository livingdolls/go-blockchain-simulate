import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

export type TTransactionFilter = {
  type: "all" | "send" | "receive";
  status: "ALL" | "PENDING" | "CONFIRMED";
  page: number;
  limit: number;
  sort_by: string;
  order: "asc" | "desc";
};

export const useTransaction = () => {
  const user = useAuthStore((state) => state.user);
  const [filter, setFilter] = useState<TTransactionFilter>({
    type: "all",
    status: "ALL",
    page: 1,
    limit: 10,
    sort_by: "created_at",
    order: "desc",
  });

  const { data, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ["transactions", filter],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not authenticated");
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
  });

  return {
    transactions: data,
    isLoading,
    isError,
    refetch,
    filter,
    setFilter,
    isFetching,
  };
};

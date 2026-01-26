import { UserBalanceRepository } from "@/repository/balance";
import { useQuery } from "@tanstack/react-query";

export const useUserBalance = (address: string) => {
  return useQuery({
    queryKey: ["user-balance", address],
    queryFn: async () => UserBalanceRepository.getUserBalance(address),
    staleTime: 60000, // 1 minute
    refetchOnWindowFocus: false,
  });
};

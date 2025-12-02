import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useQuery } from "@tanstack/react-query";

export const useTransactionNonce = () => {
  const user = useAuthStore((state) => state.user);

  return useQuery({
    queryKey: ["tx-nonce", user?.address],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not authenticated");
      }
      return await TransactionRepository.generateTxNonce(user.address);
    },
    enabled: !!user?.address,
    staleTime: 0,
    gcTime: 0,
    refetchOnMount: "always",
    refetchInterval: 120000, // 2 minutes
  });
};

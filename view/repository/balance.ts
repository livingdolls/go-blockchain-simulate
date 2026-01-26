import { api } from "@/lib/axios";
import { TUserBalance } from "@/types/balance";
import { TApiResponse } from "@/types/http";

export const UserBalanceRepository = {
  getUserBalance: async (
    address: string,
  ): Promise<TApiResponse<TUserBalance>> => {
    const response = await api.get(`/balance/${address}`);
    const data = await response.data;
    return data;
  },
};

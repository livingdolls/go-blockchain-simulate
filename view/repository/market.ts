import { api } from "@/lib/axios";
import { TMarketData } from "@/types/market";

export const MarketRepository = {
  getMarketData: async (): Promise<TMarketData> => {
    const response = await api.get("/market");

    return response.data;
  },
};

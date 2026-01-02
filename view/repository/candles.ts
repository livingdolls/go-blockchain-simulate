import { api } from "@/lib/axios";
import { TInterval, TOHLCV } from "@/types/candles";
import { TApiResponse } from "@/types/http";

export const CandlesRepository = {
  getCandlesByInterval: async (
    interval: TInterval
  ): Promise<TApiResponse<TOHLCV[]>> => {
    const response = await api.get(`/candles?interval=${interval}`);
    const data = await response.data;
    return data;
  },
};

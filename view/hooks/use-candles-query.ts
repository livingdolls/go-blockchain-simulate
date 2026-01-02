import { CandlesRepository } from "@/repository/candles";
import { TInterval } from "@/types/candles";
import { keepPreviousData, useQuery } from "@tanstack/react-query";

export const useCandlesQuery = (interval: TInterval) => {
  return useQuery({
    queryKey: ["candles", interval],
    queryFn: () => CandlesRepository.getCandlesByInterval(interval),
    staleTime: 60000, // 1 minute
    refetchOnWindowFocus: false,
    placeholderData: keepPreviousData,
  });
};

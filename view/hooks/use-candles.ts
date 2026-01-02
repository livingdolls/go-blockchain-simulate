import { TInterval } from "@/types/candles";
import { useState } from "react";
import { useCandlesQuery } from "./use-candles-query";

export const UseCandles = () => {
  const [interval, setInterval] = useState<TInterval>("1m");
  const { data, isLoading, error } = useCandlesQuery(interval);

  const changeInterval = (newInterval: TInterval) => {
    setInterval(newInterval);
  };

  return {
    candlesData: data,
    isLoading,
    error,
    interval,
    changeInterval,
  };
};

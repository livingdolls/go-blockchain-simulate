import { TInterval, TOHLCV } from "@/types/candles";
import { useEffect, useRef, useState } from "react";
import { useCandlesQuery } from "./use-candles-query";
import { SSE_BASE_URL } from "@/lib/constants/sse-url";

export const UseMarketSSE = () => {
  const [olhcData, setOhlcData] = useState<TOHLCV[]>([]);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [interval, setInterval] = useState<TInterval>("1m");

  const { data, isLoading, error } = useCandlesQuery(interval);

  // guard
  const connectionRef = useRef<EventSource | null>(null);
  const isConnectingRef = useRef<boolean>(false);

  // initialize Data from query
  useEffect(() => {
    if (data?.data && data.data.length > 0) {
      console.log("Initial candle data from query:", data.data);
      setOhlcData(data.data);
    }
  }, [data, interval]);

  useEffect(() => {
    if (isConnectingRef.current) {
      return;
    }

    isConnectingRef.current = true;
    let retryTimeout: NodeJS.Timeout;

    const connect = () => {
      if (connectionRef.current) {
        console.log("Closing existing SSE connection before reconnecting");
        connectionRef.current.close();
        connectionRef.current = null;
      }

      const url = `${SSE_BASE_URL}candles?interval=${interval}`;
      const es = new EventSource(url);
      connectionRef.current = es;

      es.onopen = () => {
        console.log("SSE connection opened");
        setIsConnected(true);
      };

      es.onmessage = (event) => {
        try {
          const newCandle: TOHLCV = JSON.parse(event.data);
          console.log("Received SSE candle data:", newCandle);

          setOhlcData((prevData) => {
            const index = prevData.findIndex(
              (candle) =>
                candle.start_time === newCandle.start_time &&
                candle.interval_type === newCandle.interval_type
            );

            if (index !== -1) {
              // Update existing candle
              console.log("Updating existing candle");
              const newData = [...prevData];
              newData[index] = newCandle;
              return newData;
            } else {
              // add  new candle and keep sorted by time
              const updated = [...prevData, newCandle].sort(
                (a, b) => a.start_time - b.start_time
              );
              return updated;
            }
          });
        } catch (error) {
          console.error("Error parsing SSE message:", error);
        }
      };

      es.onerror = (err) => {
        console.error("SSE error:", err);
        setIsConnected(false);

        if (connectionRef.current) {
          connectionRef.current.close();
          connectionRef.current = null;
        }

        retryTimeout = setTimeout(() => {
          console.log("Reconnecting to SSE...");
          connect();
        }, 3000); // retry after 5 seconds
      };
    };

    connect();

    return () => {
      console.log("Closing SSE connection");
      isConnectingRef.current = false;
      if (connectionRef.current) {
        connectionRef.current.close();
        connectionRef.current = null;
      }
      clearTimeout(retryTimeout);
      setIsConnected(false);
    };
  }, [interval]);

  const changeInterval = (newInterval: TInterval) => {
    setInterval(newInterval);
  };

  return {
    olhcData,
    isConnected,
    changeInterval,
    interval,
    isLoading,
    error,
  };
};

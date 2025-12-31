import { useEffect, useState } from "react";

export const UseMarketSSE = () => {
  const [price, setPrice] = useState(0);

  useEffect(() => {
    let es: EventSource | null = null;
    let retryTimeout: NodeJS.Timeout;

    const connect = () => {
      es = new EventSource(
        "http://192.168.88.178:3010/sse/candles?interval=1m"
      );

      es.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log("SSE data:", data);
        // Update price from candle data
        // if (data.close_price) {
        //   setPrice(data.close_price);
        // }
      };

      es.onerror = (err) => {
        console.error("SSE error:", err);
        es?.close();
        retryTimeout = setTimeout(connect, 3000); // retry 3 detik
      };
    };

    connect();

    return () => {
      es?.close();
      clearTimeout(retryTimeout);
    };
  }, []);

  return price;
};

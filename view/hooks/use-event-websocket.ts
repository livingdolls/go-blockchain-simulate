import { useCallback, useEffect, useRef, useState } from "react";

export type TEventType =
  | "market.update"
  | "block.mined"
  | "subscribe"
  | "transaction.update"
  | "balance.update";

interface IWebSocketMessage {
  type: TEventType;
  data: any;
}

interface IEventHandlers {
  [key: string]: (data: any) => void;
}
export const useEventWebSocket = (url: string) => {
  const wsRef = useRef<WebSocket | null>(null);
  const connectingRef = useRef(false); // 🔥 kunci utama
  const handlersRef = useRef<IEventHandlers>({});
  const subscribedEventsRef = useRef<TEventType[]>([]);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const connect = useCallback(() => {
    if (wsRef.current || connectingRef.current) {
      console.log("WS already active / connecting");
      return;
    }

    connectingRef.current = true;
    console.log("WS connecting...");

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("WS connected");
      connectingRef.current = false;
      setConnected(true);
      setError(null);

      if (subscribedEventsRef.current.length) {
        ws.send(
          JSON.stringify({
            type: "subscribe",
            data: {
              events: subscribedEventsRef.current,
            },
          })
        );
      }
    };

    ws.onmessage = (event) => {
      const msg: IWebSocketMessage = JSON.parse(event.data);
      console.log("WS message received:", msg);
      handlersRef.current[msg.type]?.(msg.data);
    };

    ws.onerror = (e) => {
      console.error("WS error", e);
      setError("WebSocket error");
    };

    ws.onclose = () => {
      console.log("WS closed");
      wsRef.current = null; // 🔥 WAJIB
      connectingRef.current = false; // 🔥 WAJIB
      setConnected(false);

      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    };
  }, [url]);

  useEffect(() => {
    connect();

    return () => {
      reconnectTimeoutRef.current && clearTimeout(reconnectTimeoutRef.current);

      wsRef.current?.close();
      wsRef.current = null;
      connectingRef.current = false;
    };
  }, [connect]);

  const subscribe = useCallback((events: TEventType[]) => {
    subscribedEventsRef.current = events;

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          type: "subscribe",
          data: {
            events: events,
          },
        })
      );
    }
  }, []);

  const on = useCallback((event: TEventType, handler: (data: any) => void) => {
    handlersRef.current[event] = handler;
  }, []);

  const off = useCallback((event: TEventType) => {
    delete handlersRef.current[event];
  }, []);

  return { connected, error, subscribe, on, off };
};

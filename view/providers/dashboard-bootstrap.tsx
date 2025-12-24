"use client";

import { useEventWebSocket } from "@/hooks/use-event-websocket";
import { WSEvents } from "@/lib/constants/ws-events";
import { WS_URL_MARKET } from "@/lib/constants/ws-url";
import { useDashboardStore } from "@/store/dashboard-store";
import { useEffect, useRef } from "react";

export const DashboardWSBootstrap = () => {
  const { connected, on, subscribe, error, off } =
    useEventWebSocket(WS_URL_MARKET);
  const { setConnected, setMarket, setBlock } = useDashboardStore();
  const initializedRef = useRef(false);

  // sync connection status to store
  useEffect(() => {
    setConnected(connected);
  }, [connected, setConnected]);

  // register WS handlers once
  useEffect(() => {
    if (!connected || initializedRef.current) return;

    on(WSEvents.MARKET_UPDATE, (data) => {
      console.log("Received market update:", data);
      setMarket(data);
    });

    on(WSEvents.BLOCK_MINED, (data) => {
      console.log("Received block mined event:", data);
      setBlock(data);
    });

    // subscribe to events
    subscribe([WSEvents.MARKET_UPDATE, WSEvents.BLOCK_MINED]);
    initializedRef.current = true;
  }, [connected, on, subscribe, setMarket, setBlock]);

  return null;
};

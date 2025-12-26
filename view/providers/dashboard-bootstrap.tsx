"use client";

import { useEventWebSocket } from "@/hooks/use-event-websocket";
import { WSEvents } from "@/lib/constants/ws-events";
import { WS_URL_MARKET } from "@/lib/constants/ws-url";
import { useDashboardStore } from "@/store/dashboard-store";
import { TTransactionInfo } from "@/types/transaction";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { toast } from "sonner";

export const DashboardWSBootstrap = () => {
  const { connected, on, subscribe, error, off } =
    useEventWebSocket(WS_URL_MARKET);
  const { setConnected, setMarket, setBlock } = useDashboardStore();
  const initializedRef = useRef(false);
  const qc = useQueryClient();

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

    on(WSEvents.TRANSACTION_UPDATE, (data: TTransactionInfo) => {
      toast.success(`Transaction ${data.type} of ${data.amount} confirmed!`, {
        description: `From: ${data.from_address} To: ${data.to_address}`,
      });

      qc.invalidateQueries({ queryKey: ["transactions"] });
    });

    // subscribe to events
    subscribe([
      WSEvents.MARKET_UPDATE,
      WSEvents.BLOCK_MINED,
      WSEvents.TRANSACTION_UPDATE,
    ]);
    initializedRef.current = true;
  }, [connected, on, subscribe, setMarket, setBlock]);

  return null;
};

import { TMarketData, TMinedBlockData } from "@/types/market";
import { create } from "zustand";

export type DashboardState = {
  connected: boolean;
  market: TMarketData | null;
  block: TMinedBlockData | null;

  setConnected: (connected: boolean) => void;
  setMarket: (market: TMarketData) => void;
  setBlock: (block: TMinedBlockData) => void;
};

export const useDashboardStore = create<DashboardState>((set) => ({
  connected: false,
  market: null,
  block: null,

  setConnected: (connected: boolean) =>
    set(() => ({
      connected,
    })),

  setMarket: (market: TMarketData) =>
    set(() => ({
      market,
    })),

  setBlock: (block: TMinedBlockData) =>
    set(() => ({
      block,
    })),
}));

export const WSEvents = {
  BLOCK_MINED: "block.mined",
  MARKET_UPDATE: "market.update",
} as const;

export type WSEventType = (typeof WSEvents)[keyof typeof WSEvents];

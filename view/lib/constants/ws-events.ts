export const WSEvents = {
  BLOCK_MINED: "block.mined",
  MARKET_UPDATE: "market.update",
  TRANSACTION_UPDATE: "transaction.update",
} as const;

export type WSEventType = (typeof WSEvents)[keyof typeof WSEvents];

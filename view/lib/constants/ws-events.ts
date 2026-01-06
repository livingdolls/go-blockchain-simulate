export const WSEvents = {
  BLOCK_MINED: "block.mined",
  MARKET_UPDATE: "market.update",
  TRANSACTION_UPDATE: "transaction.update",
  BALANCE_UPDATE: "balance.update",
} as const;

export type WSEventType = (typeof WSEvents)[keyof typeof WSEvents];

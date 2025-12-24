export type TMinedBlockData = {
  id: number;
  block_number: number;
  previous_hash: string;
  current_hash: string;
  nonce: number;
  difficulty: number;
  timestamp: number;
  merkle_root: string;
  miner_address: string;
  block_reward: number;
  total_fees: number;
  created_at: string;
};

export type TBlockMinedEvent = {
  type: "block.mined";
  data: TMinedBlockData;
};

// Tipe untuk data pasar
export type TMarketData = {
  id: number;
  price: number;
  liquidity: number;
  last_block: number;
  updated_at: string; // ISO 8601 format string
};

// Tipe untuk event pasar
export type TMarketUpdateEvent = {
  type: "market.update";
  data: TMarketData;
};

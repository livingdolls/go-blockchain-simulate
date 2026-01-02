export type TOHLCV = {
  id: number;
  interval_type: TInterval;
  start_time: number; // Unix timestamp in seconds
  open_price: number;
  high_price: number;
  low_price: number;
  close_price: number;
  volume: number;
};

export const TInterval = ["1m", "5m", "15m", "30m", "1h", "4h", "1d"] as const;
export type TInterval = (typeof TInterval)[number];

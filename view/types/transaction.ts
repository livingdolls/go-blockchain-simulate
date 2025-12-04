export type TTransactionWalletResponse = {
  balance: number;
  address: string;
  transactions: TTransactionsPaginate;
};

export type TTransactionsPaginate = {
  transactions: TTransactionInfo[] | null;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
};

export type TTransactionInfo = {
  id: number;
  from_address: string;
  to_address: string;
  amount: number;
  fee: number;
  signature: string;
  status: string;
  type: "send" | "received";
};

export type TTransactionFilter = {
  type: TTransactionType;
  status: TTransactionStatus;
  page: number;
  limit: number;
  sort_by: string;
  order: "asc" | "desc";
};

export const TRANSACTION_STATUSES = ["ALL", "PENDING", "CONFIRMED"] as const;
export type TTransactionStatus = (typeof TRANSACTION_STATUSES)[number];

export const TRANSACTION_TYPES = ["all", "send", "receive"] as const;
export type TTransactionType = (typeof TRANSACTION_TYPES)[number];

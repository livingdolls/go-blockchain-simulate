export type TTransactionWalletResponse = {
  balance: number;
  address: string;
  transactions: TTransactionsPaginate;
};

export type TTransactionsPaginate = {
  transactions: TTransactionInfo[];
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
  type: string;
};

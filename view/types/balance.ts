export type TSendBalance = {
  from_address: string;
  to_address: string;
  amount: number;
  private_key: string;
};

export type TSendBalanceResponse = {
  message: string;
  transaction: TTransaction;
  breakdown: TTransactionBreakdown;
  status: string;
  note: string;
};

export type TTransaction = {
  id: number;
  from_address: string;
  to_address: string;
  amount: number;
  fee: number;
  signature: string;
  status: string;
};

export type TTransactionBreakdown = {
  amount: number;
  fee: number;
  total_cost: number;
  recipient_receives: number;
};

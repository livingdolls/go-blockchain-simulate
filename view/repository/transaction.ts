import { api } from "@/lib/axios";
import { TSendBalanceResponse } from "@/types/balance";
import { TErrorAPI } from "@/types/error-api";
import {
  TTransactionStatus,
  TTransactionType,
  TTransactionWalletResponse,
} from "@/types/transaction";

export type TTransactionData = {
  from_address: string;
  to_address: string;
  amount: number;
  nonce: string;
  signature: string;
};

export type TBuySellData = {
  address: string;
  amount: number;
  nonce: string;
  signature: string;
};

export const TransactionRepository = {
  generateTxNonce: async (address: string): Promise<string> => {
    const response = await api.get<{ nonce: string }>(
      `/generate-tx-nonce/${address}`
    );
    return response.data.nonce;
  },
  sendBalance: async (
    data: TTransactionData
  ): Promise<TSendBalanceResponse | TErrorAPI> => {
    try {
      const response = await api.post<TSendBalanceResponse>("/send", data);
      return response.data;
    } catch (error: any) {
      throw new Error(error.response.data.error || "Error sending balance");
    }
  },
  getTransactionByAddress: async (
    address: string,
    page: number,
    limit: number,
    type: TTransactionType,
    status: TTransactionStatus,
    order: "asc" | "desc",
    sort_by: string
  ): Promise<TTransactionWalletResponse | TErrorAPI> => {
    try {
      const response = await api.get(`/wallet/${address}`, {
        params: {
          page,
          limit,
          type,
          status,
          order,
          sort_by,
        },
      });
      return response.data;
    } catch (error: any) {
      throw new Error(
        error.response.data.error || "Error fetching transactions"
      );
    }
  },
  buy: async (
    data: TBuySellData
  ): Promise<TSendBalanceResponse | TErrorAPI> => {
    try {
      const response = await api.post<TSendBalanceResponse>(
        "/transaction/buy",
        data
      );
      return response.data;
    } catch (error: any) {
      throw new Error(error.response.data.error || "Error processing buy");
    }
  },
  sell: async (
    data: TBuySellData
  ): Promise<TSendBalanceResponse | TErrorAPI> => {
    try {
      const response = await api.post<TSendBalanceResponse>(
        "/transaction/sell",
        data
      );
      return response.data;
    } catch (error: any) {
      throw new Error(error.response.data.error || "Error processing sell");
    }
  },
};

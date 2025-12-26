import { useEffect } from "react";
import { TransactionTable } from "./TransactionTable";
import { useTransactionQuery } from "@/hooks/use-transaction-query";

export const TransactionList = () => {
  const {
    data: transactions,
    isLoading: transactionsLoading,
    isError,
    refetch,
    isFetching,
    error,
  } = useTransactionQuery();

  useEffect(() => {
    refetch();
  }, [refetch]);

  if (transactions === undefined || isError) {
    return <div>Error loading transactions: {error?.message}</div>;
  }

  if (typeof transactions === "object" && "error" in transactions) {
    return <div>Error: {transactions.error}</div>;
  }

  return (
    <TransactionTable
      isLoading={transactionsLoading}
      isFetching={isFetching}
      data={transactions}
    />
  );
};

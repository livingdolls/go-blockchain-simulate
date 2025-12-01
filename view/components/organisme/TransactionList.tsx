import { useTransaction } from "@/hooks/use-transaction";
import { useEffect } from "react";
import { TransactionTable } from "./TransactionTable";

export const TransactionList = () => {
  const {
    transactions,
    isLoading: transactionsLoading,
    isError,
    refetch,
    filter,
    setFilter,
    isFetching,
  } = useTransaction();

  useEffect(() => {
    setFilter({ ...filter, type: "send", status: "ALL" });
  }, [filter]);

  if (transactions === undefined || isError) {
    return <div>Error loading transactions.</div>;
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

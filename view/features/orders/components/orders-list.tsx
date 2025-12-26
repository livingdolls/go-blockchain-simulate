import { TransactionTableCell } from "@/components/moleculs/transaction-tables/table-cell";
import { TransactionTableCellNotFound } from "@/components/moleculs/transaction-tables/table-cell-not-found";
import { TransactionTableHead } from "@/components/moleculs/transaction-tables/table-head";
import { TransactionTableSkeleton } from "@/components/moleculs/transaction-tables/table-skeleton";
import { PaginateTransactionTable } from "@/components/tables/paginate-transaction";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody } from "@/components/ui/table";
import { useTransactionQuery } from "@/hooks/use-transaction-query";
import { useEffect } from "react";
import { useTransactionLimitedFilter } from "../hooks/use-transaction-limited-filter";

type OrderListProps = {
  selectedTab: "buy" | "sell";
};

export const OrderList = ({ selectedTab }: OrderListProps) => {
  const {
    data: transactions,
    isLoading: transactionsLoading,
    isError,
    refetch,
    isFetching,
    error,
  } = useTransactionQuery();
  const { setType } = useTransactionLimitedFilter();

  useEffect(() => {
    refetch();
  }, [refetch]);

  useEffect(() => {
    setType("buy");
  }, []);

  useEffect(() => {
    setType(selectedTab);
  }, [selectedTab]);

  if (transactions === undefined || isError) {
    return <div>Error loading transactions: {error?.message}</div>;
  }

  if (typeof transactions === "object" && "error" in transactions) {
    return <div>Error: {transactions.error}</div>;
  }

  const headTable = [
    "ID",
    "Type",
    "Address",
    "Amount",
    "Fee",
    "Status",
    "Created At",
  ];

  if (transactionsLoading || isFetching) {
    return <TransactionTableSkeleton />;
  }

  if (!transactions || !transactions.transactions) {
    return <div>No transactions available.</div>;
  }

  return (
    <div>
      <Card>
        <CardContent>
          <Table>
            <TransactionTableHead headData={headTable} />
            <TableBody>
              {transactions.transactions.transactions === null ||
              transactions.transactions.transactions?.length === 0 ? (
                <TransactionTableCellNotFound colspan={headTable.length} />
              ) : (
                <TransactionTableCell
                  data={transactions.transactions.transactions}
                />
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          {transactions.transactions && transactions.transactions.total > 0 && (
            <PaginateTransactionTable
              total={transactions.transactions.total}
              page={transactions.transactions.page}
              limit={transactions.transactions.limit}
              total_pages={transactions.transactions.total_pages}
              isFetching={isFetching}
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
};

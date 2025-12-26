"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, RefreshCcw } from "lucide-react";
import { TTransactionWalletResponse } from "@/types/transaction";
import { TransactionTableSkeleton } from "../moleculs/transaction-tables/table-skeleton";
import { TransactionTablesIndex } from "../moleculs/transaction-tables";
import { useTransactionStore } from "@/store/transaction-store";
import { PaginateTransactionTable } from "../tables/paginate-transaction";
import { useTransactionFullFilter } from "@/hooks/use-transaction-full-filter";
import { useEffect } from "react";

type Props = {
  isLoading: boolean;
  isFetching: boolean;
  data: TTransactionWalletResponse;
};

export function TransactionTable({ isLoading, isFetching, data }: Props) {
  const { resetFilters, defaultFilter, setFilter } = useTransactionFullFilter();

  useEffect(() => {
    setFilter(defaultFilter);
    return () => {
      resetFilters();
    };
  }, []);

  if (isLoading) {
    return <TransactionTableSkeleton />;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            Transaction Sent History
            {isFetching && <Loader2 className="h-4 w-4 animate-spin" />}
          </CardTitle>
          <Button
            variant="outline"
            size="sm"
            onClick={resetFilters}
            disabled={isFetching}
          >
            <RefreshCcw className="h-4 w-4 mr-2" />
            Reset Filters
          </Button>
        </div>
      </CardHeader>

      {/* Filters */}
      <CardContent>
        <TransactionTablesIndex isLoading={isFetching} data={data} />

        {/* Pagination */}
        {data.transactions && data.transactions.total > 0 && (
          <PaginateTransactionTable
            total={data.transactions.total}
            page={data.transactions.page}
            limit={data.transactions.limit}
            total_pages={data.transactions.total_pages}
            isFetching={isFetching}
          />
        )}
      </CardContent>
    </Card>
  );
}

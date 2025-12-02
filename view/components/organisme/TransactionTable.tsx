"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Loader2,
  RefreshCcw,
} from "lucide-react";
import {
  TTransactionFilter,
  TTransactionWalletResponse,
} from "@/types/transaction";
import { TransactionTableSkeleton } from "../moleculs/transaction-tables/table-skeleton";
import { TransactionTablesIndex } from "../moleculs/transaction-tables";
import { useTransactionStore } from "@/store/transaction-store";

type Props = {
  isLoading: boolean;
  isFetching: boolean;
  data: TTransactionWalletResponse;
};

export function TransactionTable({ isLoading, isFetching, data }: Props) {
  const { goToPage, resetFilters } = useTransactionStore();

  if (isLoading) {
    return <TransactionTableSkeleton />;
  }

  const headTable = ["ID", "Type", "Address", "Amount", "Fee", "Status"];

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
        <div className="rounded-md border">
          <TransactionTablesIndex
            isLoading={isFetching}
            data={data}
            headTable={headTable}
          />
        </div>

        {/* Pagination */}
        {data.transactions && data.transactions.total > 0 && (
          <div className="flex items-center justify-between mt-4">
            <div className="text-sm text-muted-foreground">
              Showing{" "}
              <span className="font-medium">
                {(data.transactions.page - 1) * data.transactions.limit + 1}
              </span>{" "}
              to{" "}
              <span className="font-medium">
                {Math.min(
                  data.transactions.page * data.transactions.limit,
                  data.transactions.total
                )}
              </span>{" "}
              of <span className="font-medium">{data.transactions.total}</span>{" "}
              transactions
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => goToPage(1)}
                disabled={data.transactions.page === 1 || isFetching}
              >
                <ChevronsLeft className="h-4 w-4" />
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => goToPage(data.transactions.page - 1)}
                disabled={data.transactions.page === 1 || isFetching}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>

              <div className="flex items-center gap-1">
                <span className="text-sm font-medium">
                  Page {data.transactions.page} of{" "}
                  {data.transactions.total_pages}
                </span>
              </div>

              <Button
                variant="outline"
                size="sm"
                onClick={() => goToPage(data.transactions.page + 1)}
                disabled={
                  data.transactions.page === data.transactions.total_pages ||
                  isFetching
                }
              >
                <ChevronRight className="h-4 w-4" />
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => goToPage(data.transactions.total_pages)}
                disabled={
                  data.transactions.page === data.transactions.total_pages ||
                  isFetching
                }
              >
                <ChevronsRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

"use client";

import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ArrowUpRight,
  ArrowDownLeft,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Loader2,
  ArrowUpDown,
  RefreshCcw,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { TTransactionWalletResponse } from "@/types/transaction";

type Props = {
  isLoading: boolean;
  isFetching: boolean;
  data: TTransactionWalletResponse;
};

export function TransactionTable({ isLoading, isFetching, data }: Props) {
  if (isLoading) {
    return <TransactionTableSkeleton />;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            Transaction History
            {isFetching && <Loader2 className="h-4 w-4 animate-spin" />}
          </CardTitle>
          <Button
            variant="outline"
            size="sm"
            // onClick={resetFilters}
            disabled={isFetching}
          >
            <RefreshCcw className="h-4 w-4 mr-2" />
            Reset Filters
          </Button>
        </div>
      </CardHeader>

      {/* Filters */}

      <CardContent>
        {/* Table */}
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">
                  <Button variant="ghost" size="sm" className="h-8 px-2">
                    ID
                    <ArrowUpDown className="ml-2 h-4 w-4" />
                  </Button>
                </TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Address</TableHead>
                <TableHead className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 px-2 float-right"
                  >
                    Amount
                    <ArrowUpDown className="ml-2 h-4 w-4" />
                  </Button>
                </TableHead>
                <TableHead className="text-right">Fee</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {data.transactions.transactions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8">
                    <p className="text-muted-foreground">
                      No transactions found
                    </p>
                  </TableCell>
                </TableRow>
              ) : (
                data.transactions.transactions.map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell className="font-medium">#{tx.id}</TableCell>

                    <TableCell>
                      <div className="flex items-center gap-2">
                        {tx.type === "send" ? (
                          <>
                            <div className="p-1.5 bg-red-100 rounded-full">
                              <ArrowUpRight className="h-3 w-3 text-red-600" />
                            </div>
                            <span className="text-red-600 font-medium">
                              Sent
                            </span>
                          </>
                        ) : (
                          <>
                            <div className="p-1.5 bg-green-100 rounded-full">
                              <ArrowDownLeft className="h-3 w-3 text-green-600" />
                            </div>
                            <span className="text-green-600 font-medium">
                              Received
                            </span>
                          </>
                        )}
                      </div>
                    </TableCell>

                    <TableCell>
                      <div className="flex flex-col">
                        <span className="text-xs text-muted-foreground">
                          {tx.type === "send" ? "To:" : "From:"}
                        </span>
                        <code className="text-xs bg-muted px-1 py-0.5 rounded">
                          {tx.type === "send"
                            ? tx.to_address.slice(0, 10)
                            : tx.from_address.slice(0, 10)}
                          ...
                          {tx.type === "send"
                            ? tx.to_address.slice(-8)
                            : tx.from_address.slice(-8)}
                        </code>
                      </div>
                    </TableCell>

                    <TableCell
                      className={cn(
                        "text-right font-bold",
                        tx.type === "send" ? "text-red-600" : "text-green-600"
                      )}
                    >
                      {tx.type === "send" ? "-" : "+"}
                      {tx.amount.toFixed(4)} YTC
                    </TableCell>

                    <TableCell className="text-right text-muted-foreground">
                      {tx.fee.toFixed(5)} YTC
                    </TableCell>

                    <TableCell>
                      <Badge
                        variant={
                          tx.status === "CONFIRMED" ? "default" : "secondary"
                        }
                        className={
                          tx.status === "CONFIRMED"
                            ? "bg-green-500"
                            : "bg-yellow-500"
                        }
                      >
                        {tx.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
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
                disabled={data.transactions.page === 1 || isFetching}
              >
                <ChevronsLeft className="h-4 w-4" />
              </Button>

              <Button
                variant="outline"
                size="sm"
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

// Loading Skeleton
function TransactionTableSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-8 w-48" />
        <div className="flex gap-2 mt-4">
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-32" />
          <Skeleton className="h-10 w-32" />
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

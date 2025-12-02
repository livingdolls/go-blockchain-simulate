import { TableCell, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { TTransactionInfo } from "@/types/transaction";
import { ArrowDownCircle, ArrowUpCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { TableCellFilter } from "./table-cell-filter";

type TransactionTableCellProps = {
  data: TTransactionInfo[];
};

export const TransactionTableCell = ({ data }: TransactionTableCellProps) => {
  return (
    <>
      <TableCellFilter />
      {data.map((tx) => (
        <TableRow key={tx.id}>
          <TableCell className="font-medium">#{tx.id}</TableCell>

          <TableCell>
            <div className="flex items-center gap-2">
              {tx.type === "send" ? (
                <>
                  <ArrowUpCircle className="h-4 w-4 text-red-500" />
                  <span className="text-red-500">Send</span>
                </>
              ) : (
                <>
                  <ArrowDownCircle className="h-4 w-4 text-green-500" />
                  <span className="text-green-500">Receive</span>
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
            {tx.fee.toFixed(4)} YTC
          </TableCell>

          <TableCell>
            <Badge
              variant={tx.status === "CONFIRMED" ? "default" : "secondary"}
              className={
                tx.status === "CONFIRMED"
                  ? "bg-green-100 text-green-800"
                  : "bg-yellow-100 text-yellow-800"
              }
            >
              {tx.status}
            </Badge>
          </TableCell>
        </TableRow>
      ))}
    </>
  );
};

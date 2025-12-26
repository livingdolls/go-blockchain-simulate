import { TableCell, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { TTransactionInfo, TTransactionType } from "@/types/transaction";
import {
  ArrowDownCircle,
  ArrowUpCircle,
  BanknoteArrowDown,
  BanknoteArrowUp,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";

type TransactionTableCellProps = {
  data: TTransactionInfo[];
};

const typeTransaction = (type: Pick<TTransactionInfo, "type">) => {
  switch (type.type) {
    case "send":
      return (
        <>
          <ArrowUpCircle className="h-4 w-4 text-orange-500" />
          <span className="text-orange-500">Send</span>
        </>
      );
    case "received":
      return (
        <>
          <ArrowDownCircle className="h-4 w-4 text-sky-500" />
          <span className="text-sky-500">Receive</span>
        </>
      );
    case "buy":
      return (
        <>
          <BanknoteArrowUp className="h-4 w-4 text-green-500" />
          <span className="text-green-500">Buy</span>
        </>
      );
    case "sell":
      return (
        <>
          <BanknoteArrowDown className="h-4 w-4 text-red-500" />
          <span className="text-red-500">Sell</span>
        </>
      );
    default:
      return "Unknown";
  }
};

const typeTx = (type: TTransactionType, from: string, to: string) => {
  switch (type) {
    case "send":
      return (
        <>
          <span className="text-xs text-muted-foreground">To:</span>
          <code className="text-xs bg-muted px-1 py-0.5 rounded">{to}</code>
        </>
      );
    case "received":
      return (
        <>
          <span className="text-xs text-muted-foreground">From:</span>
          <code className="text-xs bg-muted px-1 py-0.5 rounded">{from}</code>
        </>
      );
    case "buy":
      return (
        <>
          <span className="text-xs text-muted-foreground">From:</span>
          <code className="text-xs bg-muted px-1 py-0.5 rounded">{from}</code>
        </>
      );
    case "sell":
      return (
        <>
          <span className="text-xs text-muted-foreground">To:</span>
          <code className="text-xs bg-muted px-1 py-0.5 rounded">{to}</code>
        </>
      );
    default:
      return "Unknown";
  }
};

export const TransactionTableCell = ({ data }: TransactionTableCellProps) => {
  return (
    <>
      {data.map((tx) => (
        <TableRow key={tx.id}>
          <TableCell className="font-medium">#{tx.id}</TableCell>

          <TableCell>
            <div className="flex items-center gap-2">
              {typeTransaction({ type: tx.type })}
            </div>
          </TableCell>

          <TableCell>
            <div className="flex flex-col">
              {typeTx(tx.type, tx.from_address, tx.to_address)}
            </div>
          </TableCell>

          <TableCell
            className={cn(
              "text-right font-bold",
              tx.type === "send" || tx.type === "sell"
                ? "text-red-600"
                : "text-green-600"
            )}
          >
            {tx.type === "send" || tx.type === "sell" ? "-" : "+"}
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

          <TableCell className="text-sm text-muted-foreground">
            {new Date(tx.created_at).toLocaleString()}
          </TableCell>
        </TableRow>
      ))}
    </>
  );
};

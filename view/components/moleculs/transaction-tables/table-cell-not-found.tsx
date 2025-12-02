import { TableCell, TableRow } from "@/components/ui/table";

type TransactionTableCellNotFoundProps = {
  colspan: number;
};

export const TransactionTableCellNotFound = ({
  colspan,
}: TransactionTableCellNotFoundProps) => {
  return (
    <TableRow>
      <TableCell colSpan={colspan} className="text-center py-8">
        <p className="text-muted-foreground">No transactions found</p>
      </TableCell>
    </TableRow>
  );
};

import { Table, TableBody } from "@/components/ui/table";
import { TransactionTableHead } from "./table-head";
import { TTransactionWalletResponse } from "@/types/transaction";
import { TransactionTableSkeleton } from "./table-skeleton";
import { TransactionTableCellNotFound } from "./table-cell-not-found";
import { TransactionTableCell } from "./table-cell";

type Props = {
  isLoading: boolean;
  data: TTransactionWalletResponse;
  headTable: string[];
};

export const TransactionTablesIndex = ({
  isLoading,
  data,
  headTable,
}: Props) => {
  if (isLoading) {
    return <TransactionTableSkeleton />;
  }

  return (
    <Table>
      <TransactionTableHead headData={headTable} />
      <TableBody>
        {data.transactions.transactions.length === 0 ? (
          <TransactionTableCellNotFound colspan={headTable.length} />
        ) : (
          <TransactionTableCell data={data.transactions.transactions} />
        )}
      </TableBody>
    </Table>
  );
};

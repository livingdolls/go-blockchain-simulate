import { Table, TableBody } from "@/components/ui/table";
import { TransactionTableHead } from "./table-head";
import { TTransactionWalletResponse } from "@/types/transaction";
import { TransactionTableSkeleton } from "./table-skeleton";
import { TransactionTableCellNotFound } from "./table-cell-not-found";
import { TransactionTableCell } from "./table-cell";
import { TableCellFilter } from "./table-cell-filter";

type Props = {
  isLoading: boolean;
  data: TTransactionWalletResponse;
};

export const TransactionTablesIndex = ({ isLoading, data }: Props) => {
  if (isLoading) {
    return <TransactionTableSkeleton />;
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

  return (
    <Table>
      <TransactionTableHead headData={headTable} />
      <TableBody>
        {data.transactions.transactions === null ||
        data.transactions.transactions.length === 0 ? (
          <TransactionTableCellNotFound colspan={headTable.length} />
        ) : (
          <>
            <TableCellFilter />
            <TransactionTableCell data={data.transactions.transactions} />
          </>
        )}
      </TableBody>
    </Table>
  );
};

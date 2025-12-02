import { TableCell, TableRow } from "@/components/ui/table";
import { TRANSACTION_STATUSES, TTransactionFilter } from "@/types/transaction";
import { SelectFilter } from "./select-filter";
import { useTransactionStore } from "@/store/transaction-store";

export const TableCellFilter = () => {
  const { updateFilter, filter: f } = useTransactionStore();

  const updateFilterData = (name: string, value: string) => {
    updateFilter({ [name]: value });
  };
  return (
    <TableRow>
      <TableCell colSpan={5}>
        <div className="py-2">Filter Transactions:</div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="space-y-2 py-2">
          <SelectFilter
            data={TRANSACTION_STATUSES as unknown as string[]}
            setFilterValue={updateFilterData}
            placeholder="Select a status"
            label="Status"
            value={f.status}
          />
        </div>
      </TableCell>
    </TableRow>
  );
};

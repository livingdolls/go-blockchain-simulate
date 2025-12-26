import { TableCell, TableRow } from "@/components/ui/table";
import { TRANSACTION_STATUSES } from "@/types/transaction";
import { SelectFilter } from "./select-filter";
import { useTransactionFullFilter } from "@/hooks/use-transaction-full-filter";

export const TableCellFilter = () => {
  const filterControl = useTransactionFullFilter();

  const updateFilterData = (name: string, value: string) => {
    filterControl.updateFilter({ [name]: value });
  };
  return (
    <TableRow>
      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="space-y-2 py-2">
          <SelectFilter
            data={TRANSACTION_STATUSES as unknown as string[]}
            setFilterValue={updateFilterData}
            placeholder="Select a status"
            label="Status"
            value={filterControl.filter.status}
          />
        </div>
      </TableCell>

      <TableCell colSpan={1}>
        <div className="py-2"></div>
      </TableCell>
    </TableRow>
  );
};

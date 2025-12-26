import { useTransactionStore } from "@/store/transaction-store";
import { TTransactionType } from "@/types/transaction";

export const useTransactionLimitedFilter = () => {
  const filter = useTransactionStore((s) => s.filter);
  const updateFilter = useTransactionStore((s) => s.updateFilter);

  const setType = (type: TTransactionType) => {
    updateFilter({ ...filter, type, page: 1 });
  };

  return {
    type: filter.type,
    setType,
  };
};

import { TableHeader, TableHead, TableRow } from "@/components/ui/table";

type TableHeadProps = {
  //   children: React.ReactNode;
  headData: string[];
};

export const TransactionTableHead = ({ headData }: TableHeadProps) => {
  return (
    <TableHeader>
      <TableRow>
        {headData.map((header) => (
          <TableHead key={header}>{header}</TableHead>
        ))}
      </TableRow>
    </TableHeader>
  );
};

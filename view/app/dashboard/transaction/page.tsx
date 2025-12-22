import { FormBuy } from "@/components/organisme/Buy/FormBuy";
import { Card } from "@/components/ui/card";

export default function TransactionPage() {
  return (
    <div className="grid grid-cols-12 gap-4">
      <Card className="p-4 col-span-12 xl:col-span-3 gap-2">
        <h2 className="mb-2 text-lg font-semibold text-center">
          Send Ballance
        </h2>
        <FormBuy />
      </Card>

      <div className="col-span-12 xl:col-span-9"></div>
    </div>
  );
}

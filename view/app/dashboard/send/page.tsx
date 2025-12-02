"use client";

import { TransactionForm } from "@/components/moleculs/transaction-form";
import { TransactionList } from "@/components/organisme/TransactionList";
import { Card } from "@/components/ui/card";

export default function SendPage() {
  return (
    <div className="grid grid-cols-12 gap-4">
      <Card className="p-4 col-span-12 xl:col-span-3 gap-2">
        <h2 className="mb-2 text-lg font-semibold text-center">
          Send Ballance
        </h2>
        <TransactionForm />
      </Card>

      <div className="col-span-12 xl:col-span-9">
        <TransactionList />
      </div>
    </div>
  );
}

"use client";

import { FormBuy } from "@/components/organisme/Buy/FormBuy";
import { FormSell } from "@/components/organisme/Sell/FormSell";
import { OrderList } from "@/features/orders/components/orders-list";
import { TabsTransaction } from "@/features/orders/components/tabs-tx";
import { useDashboardStore } from "@/store/dashboard-store";
import { useState } from "react";

export default function TransactionPage() {
  const [selectedTab, setSelectedTab] = useState<"buy" | "sell">("buy");
  const connected = useDashboardStore((state) => state.connected);
  const market = useDashboardStore((state) => state.market);
  if (!connected) {
    return <div>Connecting to WebSocket...</div>;
  }

  const handleTabChange = (tab: "buy" | "sell") => {
    setSelectedTab(tab);
  };

  return (
    <div>
      <TabsTransaction
        handleTabClick={handleTabChange}
        selectedTab={selectedTab}
      />
      <div className="grid grid-cols-12 gap-4 mt-4">
        <div className="col-span-12 lg:col-span-3">
          {selectedTab === "buy" ? <FormBuy market={market} /> : <FormSell />}
        </div>

        <div className="col-span-12 xl:col-span-9">
          <OrderList selectedTab={selectedTab} />
        </div>
      </div>
    </div>
  );
}

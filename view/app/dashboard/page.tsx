"use client";

import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { DataTable } from "@/components/data-table";
import { SectionCards } from "@/components/section-cards";
import { UseMarketSSE } from "@/hooks/use-market-sse";
import { useAuthStore } from "@/store/auth-store";
import { useDashboardStore } from "@/store/dashboard-store";

export default function Page() {
  const loading = useAuthStore((state) => state.loading);

  const { connected, market } = useDashboardStore();
  const price = UseMarketSSE();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!connected) {
    return <div>Connecting to WebSocket...</div>;
  }

  console.log("Market price from SSE:", price);

  return (
    <>
      <SectionCards />
      <div className="px-4 lg:px-6">
        <ChartAreaInteractive />
      </div>
      {/* <DataTable data={data} /> */}
    </>
  );
}

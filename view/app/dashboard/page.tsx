"use client";

import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { SectionCards } from "@/components/section-cards";
import { useUserBalance } from "@/hooks/use-user-balance";
import { useAuthStore } from "@/store/auth-store";
import { useDashboardStore } from "@/store/dashboard-store";
import { useUserBalanceStore } from "@/store/user-balance-store";

export default function Page() {
  const loading = useAuthStore((state) => state.loading);
  const user = useAuthStore((state) => state.user);
  const connected = useDashboardStore((state) => state.connected);
  const userBalance = useUserBalanceStore((state) => state.userBalance);
  const setUserBalance = useUserBalanceStore((state) => state.setUserBalance);

  const { data, isLoading } = useUserBalance(user?.address || "");

  if (data && !userBalance) {
    if (data.data !== undefined) {
      setUserBalance(data.data);
    }
  }

  if (loading || isLoading) {
    return <div>Loading...</div>;
  }

  if (!connected) {
    return <div>Connecting to WebSocket...</div>;
  }

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

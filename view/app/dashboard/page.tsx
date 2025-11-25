"use client";

import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { DataTable } from "@/components/data-table";
import { SectionCards } from "@/components/section-cards";
import { useQuery } from "@tanstack/react-query";

import data from "./data.json";

export default function Page() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["wallet-data"],
    queryFn: async () => {
      const data = await fetch(
        "http://localhost:3010/wallet/0xc1d158bfd476156099bfc1b701fb021d543fdb82"
      );
      if (!data.ok) {
        throw new Error("Network response was not ok");
      }
      return data.json();
    },
  });

  console.log("wallet data:", data, isLoading, error);

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

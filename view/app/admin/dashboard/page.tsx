"use client";

import { useAdminDashboard, useRecentActivityLogs } from "@/hooks/use-admin";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { Users, TrendingUp, Zap, AlertCircle } from "lucide-react";

export default function AdminDashboardPage() {
  const { data: dashboard, isLoading: dashboardLoading } = useAdminDashboard();
  const { data: activityLogs, isLoading: logsLoading } =
    useRecentActivityLogs(7);

  console.log("Dashboard Data:", dashboard);

  if (dashboardLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
      </div>
    );
  }

  const stats = [
    {
      label: "Total Users",
      value: dashboard?.total_users || 0,
      icon: Users,
      color: "bg-blue-500",
    },
    {
      label: "Total Admins",
      value: dashboard?.total_admins || 0,
      icon: Zap,
      color: "bg-yellow-500",
    },
    {
      label: "Total Transactions",
      value: dashboard?.total_transactions || 0,
      icon: TrendingUp,
      color: "bg-green-500",
    },
    {
      label: "Suspended Admins",
      value: dashboard?.suspended_admins || 0,
      icon: AlertCircle,
      color: "bg-red-500",
    },
  ];

  // Mock data untuk chart
  const chartData = [
    { name: "Users", value: dashboard?.total_users || 0 },
    { name: "Transactions", value: dashboard?.total_transactions || 0 },
    { name: "Blocks", value: dashboard?.total_blocks || 0 },
    { name: "Admins", value: dashboard?.total_admins || 0 },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-gray-500 mt-2">
          Ringkasan sistem dan statistik penting
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.label}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-500">{stat.label}</p>
                    <p className="text-2xl font-bold mt-2">
                      {stat.value.toLocaleString()}
                    </p>
                  </div>
                  <div className={`${stat.color} p-3 rounded-lg`}>
                    <Icon className="h-6 w-6 text-white" />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>System Overview</CardTitle>
            <CardDescription>Ringkasan data sistem</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="value" fill="#3b82f6" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Volume Trend</CardTitle>
            <CardDescription>Total volume transaksi</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center h-[300px]">
              <div className="text-center">
                <p className="text-4xl font-bold text-green-500">
                  $
                  {(dashboard?.total_volume || 0).toLocaleString("en-US", {
                    minimumFractionDigits: 2,
                  })}
                </p>
                <p className="text-gray-500 mt-2">Total Volume</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Aktivitas terbaru dari sistem</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {logsLoading ? (
              <Skeleton className="h-20" />
            ) : activityLogs && activityLogs.length > 0 ? (
              activityLogs.slice(0, 5).map((log) => (
                <div
                  key={log.id}
                  className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50"
                >
                  <div className="flex-1">
                    <p className="font-medium">
                      {log.action.toUpperCase()} - {log.target_entity}
                    </p>
                    <p className="text-sm text-gray-500">
                      {log.changes_summary}
                    </p>
                  </div>
                  <div className="text-right">
                    <p
                      className={`text-sm font-medium ${
                        log.status === "success"
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {log.status.toUpperCase()}
                    </p>
                    <p className="text-xs text-gray-500">
                      {new Date(log.created_at).toLocaleDateString("id-ID")}
                    </p>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-center text-gray-500">
                Tidak ada aktivitas terbaru
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

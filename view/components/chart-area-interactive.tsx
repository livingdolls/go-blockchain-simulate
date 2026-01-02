"use client";

import * as React from "react";
import {
  ComposedChart,
  Bar,
  Line,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

import { useIsMobile } from "@/hooks/use-mobile";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { UseMarketSSE } from "@/hooks/use-market-sse";
import { TInterval } from "@/types/candles";

export const description = "An interactive OHLCV candlestick chart";

interface CandleData {
  id: number;
  interval_type: string;
  start_time: number;
  open_price: number;
  high_price: number;
  low_price: number;
  close_price: number;
  volume: number;
}

const chartConfig = {
  open: {
    label: "Open",
    color: "var(--primary)",
  },
  high: {
    label: "High",
    color: "#10b981",
  },
  low: {
    label: "Low",
    color: "#ef4444",
  },
  close: {
    label: "Close",
    color: "#3b82f6",
  },
  volume: {
    label: "Volume",
    color: "#8b5cf6",
  },
} satisfies ChartConfig;

export function ChartAreaInteractive() {
  const { olhcData, isConnected, interval, changeInterval, isLoading, error } =
    UseMarketSSE();

  // Format data untuk chart
  const chartData = React.useMemo(() => {
    return (olhcData || []).map((candle) => ({
      time: new Date(candle.start_time * 1000).toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      }),
      timestamp: candle.start_time,
      open: parseFloat(candle.open_price.toFixed(2)),
      high: parseFloat(candle.high_price.toFixed(2)),
      low: parseFloat(candle.low_price.toFixed(2)),
      close: parseFloat(candle.close_price.toFixed(2)),
      volume: parseFloat(candle.volume.toFixed(2)),
    }));
  }, [olhcData]);

  // Get latest candle stats
  const latestCandle = React.useMemo(() => {
    if (chartData.length === 0) return null;
    return chartData[chartData.length - 1];
  }, [chartData]);

  return (
    <Card className="@container/card">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Candlestick Chart ({interval})</CardTitle>
            <CardDescription>
              <span className="hidden @[540px]/card:block">
                Real-time OHLCV data
              </span>
              <span className="@[540px]/card:hidden">OHLCV Data</span>
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`inline-block w-2 h-2 rounded-full ${
                isConnected ? "bg-green-500" : "bg-red-500"
              }`}
            />
            <span className="text-sm text-gray-500">
              {isConnected ? "Live" : "Offline"}
            </span>
          </div>
        </div>
        <CardAction>
          <Select value={interval} onValueChange={changeInterval}>
            <SelectTrigger
              className="flex w-32 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate"
              size="sm"
            >
              <SelectValue placeholder="Select interval" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              {TInterval.map((value) => (
                <SelectItem key={value} value={value} className="rounded-lg">
                  {value}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardAction>
      </CardHeader>

      {/* Latest Candle Stats */}
      {latestCandle && (
        <div className="px-6 pt-4">
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-5 mb-4">
            <div className="bg-gray-50 dark:bg-gray-900 p-3 rounded-lg">
              <p className="text-xs font-medium text-gray-600 dark:text-gray-400">
                Open
              </p>
              <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
                ${latestCandle.open.toFixed(4)}
              </p>
            </div>
            <div className="bg-green-50 dark:bg-green-900/20 p-3 rounded-lg">
              <p className="text-xs font-medium text-green-600 dark:text-green-400">
                High
              </p>
              <p className="text-lg font-bold text-green-600 dark:text-green-400">
                ${latestCandle.high.toFixed(4)}
              </p>
            </div>
            <div className="bg-red-50 dark:bg-red-900/20 p-3 rounded-lg">
              <p className="text-xs font-medium text-red-600 dark:text-red-400">
                Low
              </p>
              <p className="text-lg font-bold text-red-600 dark:text-red-400">
                ${latestCandle.low.toFixed(4)}
              </p>
            </div>
            <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
              <p className="text-xs font-medium text-blue-600 dark:text-blue-400">
                Close
              </p>
              <p className="text-lg font-bold text-blue-600 dark:text-blue-400">
                ${latestCandle.close.toFixed(4)}
              </p>
            </div>
            <div className="bg-purple-50 dark:bg-purple-900/20 p-3 rounded-lg">
              <p className="text-xs font-medium text-purple-600 dark:text-purple-400">
                Volume
              </p>
              <p className="text-lg font-bold text-purple-600 dark:text-purple-400">
                {latestCandle.volume.toFixed(2)}
              </p>
            </div>
          </div>
        </div>
      )}

      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <div className="w-full h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart
              data={chartData}
              margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="time" tick={{ fontSize: 12 }} tickMargin={8} />
              <YAxis
                yAxisId="left"
                tick={{ fontSize: 12 }}
                domain={["dataMin - 0.1", "dataMax + 0.1"]}
              />
              <YAxis
                yAxisId="right"
                orientation="right"
                tick={{ fontSize: 12 }}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "rgba(0, 0, 0, 0.8)",
                  border: "1px solid #666",
                  borderRadius: "8px",
                }}
                formatter={(value: any) => {
                  if (typeof value === "number") {
                    return [value.toFixed(4), ""];
                  }
                  return value;
                }}
                labelFormatter={(label) => `Time: ${label}`}
              />
              <Legend
                verticalAlign="top"
                height={36}
                wrapperStyle={{ paddingBottom: "10px" }}
              />
              {/* OHLC Lines */}
              <Line
                yAxisId="left"
                type="monotone"
                dataKey="open"
                stroke="#3b82f6"
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="left"
                type="monotone"
                dataKey="high"
                stroke="#10b981"
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="left"
                type="monotone"
                dataKey="low"
                stroke="#ef4444"
                dot={false}
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="left"
                type="monotone"
                dataKey="close"
                stroke="#8b5cf6"
                dot={false}
                strokeWidth={2}
                name="Close"
                isAnimationActive={false}
              />
              {/* Volume Bar */}
              <Bar
                yAxisId="right"
                dataKey="volume"
                fill="#fbbf24"
                opacity={0.3}
                name="Volume"
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      </CardContent>

      {/* Data Table */}
      <div className="px-6 pb-6">
        <h3 className="text-sm font-semibold mb-3">Recent Candles</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead className="border-b bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-2 py-2 text-left font-medium">Time</th>
                <th className="px-2 py-2 text-right font-medium">Open</th>
                <th className="px-2 py-2 text-right font-medium">High</th>
                <th className="px-2 py-2 text-right font-medium">Low</th>
                <th className="px-2 py-2 text-right font-medium">Close</th>
                <th className="px-2 py-2 text-right font-medium">Volume</th>
              </tr>
            </thead>
            <tbody>
              {chartData
                .slice(-10)
                .reverse()
                .map((row, idx) => (
                  <tr
                    key={idx}
                    className="border-b hover:bg-gray-50 dark:hover:bg-gray-900"
                  >
                    <td className="px-2 py-2">{row.time}</td>
                    <td className="px-2 py-2 text-right">
                      ${row.open.toFixed(4)}
                    </td>
                    <td className="px-2 py-2 text-right text-green-600">
                      ${row.high.toFixed(4)}
                    </td>
                    <td className="px-2 py-2 text-right text-red-600">
                      ${row.low.toFixed(4)}
                    </td>
                    <td className="px-2 py-2 text-right font-medium">
                      ${row.close.toFixed(4)}
                    </td>
                    <td className="px-2 py-2 text-right text-purple-600">
                      {row.volume.toFixed(2)}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
        <p className="text-xs text-gray-500 mt-3">
          Showing last 10 candles • Total: {chartData.length} candles
        </p>
      </div>
    </Card>
  );
}

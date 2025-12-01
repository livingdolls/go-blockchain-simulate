"use client";

import { TransactionList } from "@/components/organisme/TransactionList";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useSendBalance } from "@/hooks/use-send-balance";
import { useTransaction } from "@/hooks/use-transaction";
import { useTransactionNonce } from "@/hooks/use-transaction-nonce";
import { Loader2 } from "lucide-react";
import { useState } from "react";

export default function SendPage() {
  const [toAddress, setToAddress] = useState("");
  const [amount, setAmount] = useState(0);
  const [mnemonic, setMnemonic] = useState("");

  const {
    data: nonceData,
    isLoading: nonceLoading,
    refetch: refetchNonce,
  } = useTransactionNonce();

  const { sendTransaction, isLoading } = useSendBalance();

  const resetForm = () => {
    setToAddress("");
    setAmount(0);
    setMnemonic("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nonceData) {
      console.error("Nonce data is not available");
      return;
    }
    try {
      await sendTransaction(toAddress, amount, nonceData, mnemonic);
      console.log("Transaction sent successfully");
      resetForm();
      refetchNonce();
    } catch (err) {
      console.error("Error sending transaction:", err);
    }
  };

  if (nonceLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="grid grid-cols-12 gap-4">
      {/* Card Send */}
      <Card className="p-4 col-span-12 xl:col-span-3 gap-2">
        <h2 className="mb-2 text-lg font-semibold text-center">
          Send Cryptocurrency
        </h2>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col">
            <label htmlFor="recipient" className="mb-1 font-medium">
              Recipient Address
            </label>
            <Input
              type="text"
              id="recipient"
              className="rounded border border-gray-300 p-2"
              placeholder="Enter recipient address"
              onChange={(e) => setToAddress(e.target.value)}
              value={toAddress}
            />
          </div>
          <div className="flex flex-col">
            <label htmlFor="amount" className="mb-1 font-medium">
              Amount
            </label>
            <Input
              type="number"
              id="amount"
              className="rounded border border-gray-300 p-2"
              placeholder="Enter amount to send"
              onChange={(e) => setAmount(Number(e.target.value))}
              value={amount}
            />
          </div>

          <div className="flex flex-col">
            <label htmlFor="mnemonic" className="mb-1 font-medium">
              Mnemonic Key
            </label>
            <Textarea
              id="mnemonic"
              className="rounded border border-gray-300 p-2"
              placeholder="Enter your mnemonic key"
              onChange={(e) => setMnemonic(e.target.value)}
              value={mnemonic}
              defaultValue={
                "spatial media crunch crop clump candy rotate hollow amount tissue total scene"
              }
              //fancy pair mammal swarm they syrup discover school rug obtain extend hotel
              rows={3}
            />
          </div>

          <button
            type="submit"
            className="mt-4 rounded bg-black px-4 py-2 text-white "
            disabled={isLoading}
          >
            Send
            {isLoading && <Loader2 className="animate-spin" />}
          </button>
        </form>
      </Card>

      {/* Ini nanti di isi chart, dan riwayat send transaksi */}
      <div className="col-span-12 xl:col-span-9">
        <TransactionList />
      </div>
    </div>
  );
}

"use client";

import { WalletFileDropzone } from "@/components/moleculs/transaction-form/dropzone";
import { InputTx } from "@/components/moleculs/transaction-form/input-tx";
import { TextareaTx } from "@/components/moleculs/transaction-form/textarea-tx";
import { Card } from "@/components/ui/card";
import { FieldSeparator } from "@/components/ui/field";
import { useBuy } from "@/hooks/use-buy";
import { useTransactionNonce } from "@/hooks/use-transaction-nonce";
import { TMarketData } from "@/types/market";
import { toast } from "sonner";

type props = {
  market: TMarketData | null;
};

export const FormBuy = ({ market }: props) => {
  const { fileWallet, handleWalletFileChange, form, handleChange, executeBuy } =
    useBuy();

  const {
    data: nonceData,
    isLoading: nonceLoading,
    refetch: refetchNonce,
  } = useTransactionNonce();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!nonceData) {
      toast.error("Nonce data is not available, please try again");
      refetchNonce();
      return;
    }

    await executeBuy(nonceData);
    refetchNonce();
  };

  if (nonceLoading) {
    return <div>Loading...</div>;
  }

  return (
    <Card className="p-4">
      <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
        <InputTx
          label="Amount"
          type="number"
          name="amount"
          placeholder="Enter amount to buy"
          onChange={handleChange}
          value={form.amount}
          disabled={false}
        />

        {fileWallet === null && (
          <TextareaTx
            label="Mnemonic"
            name="mnemonic"
            placeholder="Enter your mnemonic phrase"
            onChange={handleChange}
            value={form.mnemonic}
            disabled={false}
          />
        )}

        <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card my-6">
          Or use your wallet file
        </FieldSeparator>

        <WalletFileDropzone
          onFile={(file, content) => {
            handleWalletFileChange(file, content);
          }}
        />

        {fileWallet !== null && (
          <InputTx
            label="Password"
            name="password"
            type="password"
            placeholder="Enter your password"
            onChange={handleChange}
            value={form.password}
            disabled={false}
          />
        )}

        <div>
          <p className="text-sm">
            Current Price: {market ? market.price.toFixed(2) : ""} USD per YTC
          </p>
          <p className="text-sm">
            <strong>Amount to pay:</strong>{" "}
            {form.amount ? form.amount.toFixed(2) : ""} USD
          </p>
          <p className="text-sm">
            <strong>YTC to receive:</strong>{" "}
            {form.amount && market
              ? (form.amount / market.price).toFixed(8)
              : ""}{" "}
            YTC
          </p>
        </div>

        <button
          type="submit"
          className="mt-4 cursor-pointer rounded bg-black px-4 py-2 text-white flex flex-row items-center gap-2 justify-center disabled:opacity-50"
        >
          Buy
        </button>
      </form>
    </Card>
  );
};

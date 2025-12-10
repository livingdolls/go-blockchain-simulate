import { InputTx } from "./input-tx";
import { useTransactionNonce } from "@/hooks/use-transaction-nonce";
import { useSendBalance } from "@/hooks/use-send-balance";
import { TextareaTx } from "./textarea-tx";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { FieldSeparator } from "@/components/ui/field";
import { WalletFileDropzone } from "./dropzone";

export const TransactionForm = () => {
  const {
    data: nonceData,
    isLoading: nonceLoading,
    refetch: refetchNonce,
  } = useTransactionNonce();

  const {
    sendTransaction,
    isLoading,
    form,
    handleChange,
    fileWallet,
    handleWalletFileChange,
  } = useSendBalance();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!nonceData) {
      toast.error("Nonce data is not available");
      return;
    }
    await sendTransaction(nonceData);
    refetchNonce();
  };

  if (nonceLoading) {
    return <div>Loading...</div>;
  }
  return (
    <>
      <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
        <InputTx
          label="Recipient Address"
          name="toAddress"
          type="text"
          placeholder="Enter recipient address"
          onChange={handleChange}
          value={form.toAddress}
          disabled={isLoading}
        />
        <InputTx
          label="Amount"
          name="amount"
          type="number"
          placeholder="Enter amount to send"
          onChange={handleChange}
          value={form.amount}
          disabled={isLoading}
        />

        {fileWallet === null && (
          <TextareaTx
            label="Mnemonic Key"
            name="mnemonic"
            placeholder="Enter your mnemonic key"
            onChange={handleChange}
            value={form.mnemonic}
            disabled={isLoading}
          />
        )}

        <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card my-6">
          Or use your wallet file
        </FieldSeparator>

        <WalletFileDropzone
          onFile={(file, content) => {
            handleWalletFileChange(file, content);
          }}
          disabled={isLoading}
        />

        {fileWallet !== null && (
          <InputTx
            label="Password"
            name="password"
            type="password"
            placeholder="Enter your password"
            onChange={handleChange}
            value={form.password}
            disabled={isLoading}
          />
        )}

        <button
          type="submit"
          className="mt-4 cursor-pointer rounded bg-black px-4 py-2 text-white flex flex-row items-center gap-2 justify-center disabled:opacity-50"
          disabled={isLoading}
        >
          Send
          {isLoading && <Loader2 className="animate-spin" />}
        </button>
      </form>
    </>
  );
};

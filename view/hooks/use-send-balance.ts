import { WalletFromMnemonic } from "@/lib/crypto";
import { walletRestoreFromBackup } from "@/lib/wallet-backup";
import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { verifyMessage } from "ethers/hash";
import { useState } from "react";
import { toast } from "sonner";

export const useSendBalance = () => {
  const qc = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const [form, setForm] = useState({
    toAddress: "",
    amount: 0,
    mnemonic: "",
    password: "",
  });

  const [fileWallet, setFileWallet] = useState<File | null>(null);

  const sendBalanceMutation = useMutation({
    mutationFn: TransactionRepository.sendBalance,
    onError: (error) => {
      toast.error(error.message || "Failed to send balance");
    },
    onSuccess: () => {
      toast.success("Balance sent successfully");
      sendBalanceMutation.reset();
      resetForm();
      qc.invalidateQueries({ queryKey: ["transactions"] });
    },
  });

  const resetForm = () => {
    setForm({
      toAddress: "",
      amount: 0,
      mnemonic: "",
      password: "",
    });
  };

  const sendTransaction = async (nonce: string) => {
    if (!user) {
      toast.error("User not authenticated");
      return;
    }

    if (fileWallet) {
      return new Promise<boolean>((resolve) => {
        const reader = new FileReader();
        reader.onload = async (e) => {
          try {
            const content = e.target?.result;

            if (typeof content !== "string") {
              toast.error("Invalid file content.");
              resolve(false);
              return;
            }

            const validate = await walletRestoreFromBackup(
              form.password,
              content
            );

            if (!validate.ok || !validate.wallet) {
              toast.error("Failed to restore wallet: " + validate.error);
              resolve(false);
              return;
            }

            const wallet = validate.wallet;
            const fromAddress = wallet.address.toLowerCase();

            // validasii user address
            if (fromAddress !== user.address.toLowerCase()) {
              toast.error(
                "Derived address does not match authenticated user address"
              );
              resolve(false);
              return;
            }

            const normalizedToAddress = form.toAddress.toLowerCase();

            const formattedAmount = form.amount.toFixed(2);
            const message = `Send ${formattedAmount} to ${normalizedToAddress} nonce:${nonce}`;

            //sign message
            const signature = await wallet.signMessage(message);

            // verify localy signature
            const recovered = verifyMessage(message, signature).toLowerCase();

            if (recovered !== fromAddress) {
              toast.error("Signature verification failed");
              resolve(false);
              return;
            }

            // send transaction
            const txData = {
              from_address: fromAddress,
              to_address: normalizedToAddress,
              amount: parseFloat(formattedAmount),
              nonce,
              signature,
            };

            await sendBalanceMutation.mutateAsync(txData);
            resolve(true);
          } catch (error) {
            toast.error("Invalid wallet file.");
            resolve(false);
          }
        };
        reader.readAsText(fileWallet);
      });
    }

    // derive wallet from mnemonic
    const wallet = WalletFromMnemonic(form.mnemonic);
    const fromAddress = wallet.address.toLowerCase();

    // validasii user address
    if (fromAddress !== user.address.toLowerCase()) {
      toast.error("Derived address does not match authenticated user address");
      return;
    }

    const normalizedToAddress = form.toAddress.toLowerCase();

    const formattedAmount = form.amount.toFixed(2);
    const message = `Send ${formattedAmount} to ${normalizedToAddress} nonce:${nonce}`;

    //sign message
    const signature = await wallet.signMessage(message);

    // verify localy signature
    const recovered = verifyMessage(message, signature).toLowerCase();

    if (recovered !== fromAddress) {
      toast.error("Signature verification failed");
      return;
    }

    // send transaction
    const txData = {
      from_address: fromAddress,
      to_address: normalizedToAddress,
      amount: parseFloat(formattedAmount),
      nonce,
      signature,
    };

    await sendBalanceMutation.mutateAsync(txData);
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setForm((prev) => ({
      ...prev,
      [name]: name === "amount" ? Number(value) : value,
    }));
  };

  const handleWalletFileChange = (file: File | null, content: unknown) => {
    if (content && typeof content !== "object") {
      toast.error("Invalid wallet file content");
      return;
    }

    setFileWallet(file);
  };

  return {
    sendTransaction,
    isLoading: sendBalanceMutation.isPending,
    form,
    handleChange,
    fileWallet,
    handleWalletFileChange,
  };
};

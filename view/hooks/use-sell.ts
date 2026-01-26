import { ValidateWallet } from "@/lib/validate-wallet";
import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { verifyMessage } from "ethers/hash";
import { useState } from "react";
import { toast } from "sonner";

export const useSell = () => {
  const qc = useQueryClient();
  const [fileWallet, setFileWallet] = useState<File | null>(null);
  const [form, setForm] = useState({
    amount: 0,
    mnemonic: "",
    password: "",
  });
  const user = useAuthStore((state) => state.user);

  const handleWalletFileChange = (file: File | null, content: unknown) => {
    if (content && typeof content !== "object") {
      toast.error("Invalid wallet file content");
      return;
    }
    setFileWallet(file);
  };

  const sellMutation = useMutation({
    mutationFn: TransactionRepository.sell,
    onError: (error) => {
      toast.error(error.message || "Failed to execute sell");
    },
    onSuccess: () => {
      toast.success("Sell executed successfully");
      sellMutation.reset();
      resetForm();
      qc.invalidateQueries({ queryKey: ["transactions"] });
    },
  });

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const { name, value } = e.target;
    setForm((prev) => ({
      ...prev,
      [name]: name === "amount" ? Number(value) : value,
    }));
  };

  const executeSell = async (nonce: string) => {
    if (!user) {
      toast.error("User not authenticated");
      return;
    }

    if (fileWallet) {
      const validate = await ValidateWallet(fileWallet, form.password);

      if (!validate.ok || !validate.wallet) {
        toast.error(validate.error || "Invalid wallet or password");
        return;
      }

      const wallet = validate.wallet;
      const address = wallet.address.toLowerCase();

      if (address !== user.address.toLowerCase()) {
        toast.error(
          "Derived address does not match authenticated user address",
        );
        return;
      }

      const formattedAmount = form.amount.toFixed(2);
      const message = ` SELL ${formattedAmount} nonce:${nonce}`;

      // sign message
      const signature = await wallet.signMessage(message);

      // verify locally signature
      const recovered = verifyMessage(message, signature).toLowerCase();

      if (recovered !== address) {
        toast.error("Signature verification failed");
        return;
      }

      const txData = {
        address: address,
        amount: parseFloat(formattedAmount),
        nonce,
        signature,
      };

      await sellMutation.mutateAsync(txData);

      return;
    }
  };

  const resetForm = () => {
    setForm({
      amount: 0,
      mnemonic: "",
      password: "",
    });
    setFileWallet(null);
  };

  return {
    fileWallet,
    handleWalletFileChange,
    form,
    handleChange,
    executeSell,
  };
};

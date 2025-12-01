import { WalletFromMnemonic } from "@/lib/crypto";
import { TransactionRepository } from "@/repository/transaction";
import { useAuthStore } from "@/store/auth-store";
import { useMutation } from "@tanstack/react-query";
import { verifyMessage } from "ethers/hash";

export const useSendBalance = () => {
  const user = useAuthStore((state) => state.user);

  const sendBalanceMutation = useMutation({
    mutationFn: TransactionRepository.sendBalance,
    onError: (error) => {
      console.error("Error sending balance:", error);
    },
    onSuccess: (data) => {
      console.log("Balance sent successfully:", data);
      sendBalanceMutation.reset();
    },
  });

  const sendTransaction = async (
    toAddress: string,
    amount: number,
    nonce: string,
    mnemonic: string
  ) => {
    if (!user) {
      throw new Error("User not authenticated");
    }

    // derive wallet from mnemonic
    const wallet = WalletFromMnemonic(mnemonic);
    const fromAddress = wallet.address.toLowerCase();

    // validasii user address
    if (fromAddress !== user.address.toLowerCase()) {
      throw new Error(
        "Derived address does not match authenticated user address"
      );
    }

    const normalizedToAddress = toAddress.toLowerCase();

    const formattedAmount = amount.toFixed(2);
    const message = `Send ${formattedAmount} to ${normalizedToAddress} nonce:${nonce}`;

    //sign message
    const signature = await wallet.signMessage(message);

    // verify localy signature
    const recovered = verifyMessage(message, signature).toLowerCase();

    if (recovered !== fromAddress) {
      throw new Error("Signature verification failed");
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

  return {
    sendTransaction,
    isLoading: sendBalanceMutation.isPending,
  };
};

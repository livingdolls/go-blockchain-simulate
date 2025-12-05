import { useState } from "react";
import { hashMessage, verifyMessage } from "ethers";
import { toast } from "sonner";
import { WalletFromMnemonic } from "@/lib/crypto";
import { walletRestoreFromBackup } from "@/lib/wallet-backup";
import { api } from "@/lib/axios";

export const useLogin = () => {
  const [mnemonic, setMnemonic] = useState("");
  const [password, setPassword] = useState("");
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const togglePasswordVisibility = () => {
    setIsPasswordVisible((prev) => !prev);
  };

  const loginWithWallet = async (walletAddress: string, wallet: any) => {
    try {
      const addr = walletAddress.toLowerCase();

      // Get challenge from server
      const ch = await api.post(`/challenge/${addr}`);
      const nonce = ch.data.challenge;
      if (!nonce) {
        toast.error("Failed to get challenge from server.");
        return false;
      }

      // Sign canonical message
      const message = `Login to YuteBlockchain nonce:${nonce}`;
      const signature = await wallet.signMessage(message);

      // Verify signature locally
      const recovered = verifyMessage(message, signature);
      if (recovered.toLowerCase() !== addr) {
        toast.error("Signature verification failed.");
        return false;
      }

      // Send signature to server for verification
      const payload = {
        address: addr,
        signature,
        nonce,
      };

      const res = await api.post(`/challenge/verify`, payload);

      if (res.data.valid) {
        window.location.href = "/dashboard";
        return true;
      } else {
        toast.error("Login failed. Invalid signature.");
        return false;
      }
    } catch (error) {
      console.error("Login error:", error);
      toast.error("Login failed. Please try again.");
      return false;
    }
  };

  const handleLoginWithMnemonic = async () => {
    if (!mnemonic) {
      toast.error("Please enter your mnemonic.");
      return false;
    }

    try {
      const wallet = WalletFromMnemonic(mnemonic);
      return await loginWithWallet(wallet.address, wallet);
    } catch (error) {
      console.error("Mnemonic login error:", error);
      toast.error("Invalid mnemonic phrase.");
      return false;
    }
  };

  const handleLoginWithFile = async () => {
    if (!file) {
      toast.error("Please select a wallet file.");
      return false;
    }

    if (!password) {
      toast.error("Please enter your password.");
      return false;
    }

    return new Promise<boolean>((resolve) => {
      const reader = new FileReader();
      reader.onload = async (event) => {
        try {
          const content = event.target?.result;
          if (typeof content !== "string") {
            toast.error("Invalid file content.");
            resolve(false);
            return;
          }

          const validate = await walletRestoreFromBackup(password, content);
          if (!validate.ok || !validate.wallet) {
            toast.error("Failed to restore wallet: " + validate.error);
            resolve(false);
            return;
          }

          const wallet = validate.wallet;
          const success = await loginWithWallet(wallet.address, wallet);
          resolve(success);
        } catch (error) {
          console.error("File login error:", error);
          toast.error("Failed to restore wallet from file.");
          resolve(false);
        }
      };

      reader.onerror = () => {
        toast.error("Failed to read file.");
        resolve(false);
      };

      reader.readAsText(file);
    });
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      if (file) {
        await handleLoginWithFile();
      } else {
        await handleLoginWithMnemonic();
      }
    } finally {
      setIsLoading(false);
    }
  };

  const resetForm = () => {
    setMnemonic("");
    setPassword("");
    setFile(null);
    setIsPasswordVisible(false);
  };

  return {
    // State
    mnemonic,
    password,
    isPasswordVisible,
    file,
    isLoading,

    // Setters
    setMnemonic,
    setPassword,
    setFile,

    // Actions
    handleLogin,
    togglePasswordVisibility,
    resetForm,
  };
};

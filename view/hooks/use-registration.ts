import { CreateWallet } from "@/lib/crypto";
import { TRegister, TRegisterResponse } from "@/types/register";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Register } from "@/repository/register";
import { createEthersBackup } from "@/lib/wallet-backup";
import { toast } from "sonner";
import { DownloadFile } from "@/lib/download-filte";
import { TApiResponse } from "@/types/http";

export type MnemonicWallet = {
  mnemonic: string;
  publicKey: string;
  privateKey: string;
  address: string;
};

export function useRegistration() {
  const [currentStep, setCurrentStep] = useState(1);
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [repeatPassword, setRepeatPassword] = useState<string>("");

  const [wallet, setWallet] = useState<MnemonicWallet>({
    mnemonic: "",
    publicKey: "",
    privateKey: "",
    address: "",
  });

  const nextStep = () => {
    setCurrentStep((prev) => prev + 1);
  };

  const prevStep = () => {
    setCurrentStep((prev) => Math.max(prev - 1, 1));
  };

  const generateWallet = async () => {
    const wallet = await CreateWallet();
    setWallet({
      mnemonic: wallet.mnemonic,
      publicKey: wallet.publicKey,
      privateKey: wallet.privateKey,
      address: wallet.address,
    });
  };

  const mutateRegister = useMutation<
    TApiResponse<TRegisterResponse>,
    Error,
    TRegister
  >({
    mutationFn: Register,
    onSuccess: (data) => {
      if (data.success) {
        toast.success(
          "Registration successful!, please wait... we are redirecting you."
        );
        window.location.href = "/dashboard";
      } else {
        toast.error(`Registration failed: ${data.error}`);
      }
    },
    onError: (error) => {
      toast.error(`Registration failed: ${error.message}`);
    },
  });

  const handleSubmitRegistration = (e: React.FormEvent) => {
    e.preventDefault();
    const data: TRegister = {
      username: username,
      address: wallet.address,
      public_key: wallet.publicKey,
    };

    mutateRegister.mutate(data);
  };

  const onChangeForm = (
    e: React.ChangeEvent<HTMLInputElement>,
    field: string
  ) => {
    if (field === "password") setPassword(e.target.value);
    if (field === "repeatPassword") setRepeatPassword(e.target.value);
    if (field === "username") setUsername(e.target.value);
  };

  const handleDownloadBackup = async () => {
    if (!wallet.privateKey) {
      toast.error("Wallet private key is missing.");
      return;
    }

    if (!password) {
      toast.error("Password is required to create wallet backup.");
      return;
    }

    if (password !== repeatPassword) {
      toast.error("Passwords do not match.");
      return;
    }

    const keystore = await createEthersBackup(wallet.privateKey, password);

    DownloadFile(
      keystore,
      `wallet-backup-${wallet.address.substring(2, 10)}.json`
    );

    toast.success("Wallet backup downloaded successfully.");
  };

  return {
    currentStep,
    nextStep,
    prevStep,
    generateWallet,
    wallet,
    handleSubmitRegistration,
    onChangeForm,
    username,
    password,
    repeatPassword,
    handleDownloadBackup,
  };
}

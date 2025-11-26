import { CreateWallet } from "@/lib/crypto";
import { TRegister, TRegisterResponse } from "@/types/register";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Register } from "@/repository/register";

export type MnemonicWallet = {
  mnemonic: string;
  publicKey: string;
  privateKey: string;
  address: string;
};

export function useRegistration() {
  const [currentStep, setCurrentStep] = useState(1);
  const [username, setUsername] = useState<string>("");
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

  const mutateRegister = useMutation<TRegisterResponse, Error, TRegister>({
    mutationFn: Register,
    onSuccess: (data) => {
      console.log("Registration successful:", data);
      window.location.href = "/dashboard";
    },
    onError: (error) => {
      console.error("Registration failed:", error);
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

  const onChangeUsername = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUsername(e.target.value);
  };

  return {
    currentStep,
    nextStep,
    prevStep,
    generateWallet,
    wallet,
    handleSubmitRegistration,
    onChangeUsername,
    username,
  };
}

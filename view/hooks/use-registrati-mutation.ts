import { Register } from "@/repository/register";
import { TRegister, TRegisterResponse } from "@/types/register";
import { useMutation } from "@tanstack/react-query";

export const useRegistratiMutation = (data: TRegister) => {
  return useMutation<TRegisterResponse, Error, TRegister>({
    mutationFn: () => Register(data),
    onSuccess: (data) => {
      console.log("Registration successful:", data);
    },
    onError: (error) => {
      console.error("Registration failed:", error);
    },
  });
};

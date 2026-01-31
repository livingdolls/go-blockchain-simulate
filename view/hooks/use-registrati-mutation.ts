import { Register } from "@/repository/register";
import { TApiResponse } from "@/types/http";
import { TRegister, TRegisterResponse } from "@/types/register";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

export const useRegistratiMutation = (data: TRegister) => {
  return useMutation<TApiResponse<TRegisterResponse>, Error, TRegister>({
    mutationFn: () => Register(data),
    onSuccess: (data) => {
      toast.success("Registration successfull, please wait...")
    },
    onError: (error) => {
      toast.error("failed to login, please check username or password!")
    },
  });
};

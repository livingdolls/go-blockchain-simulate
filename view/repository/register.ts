import { api } from "@/lib/axios";
import { TRegister, TRegisterResponse } from "@/types/register";

export const Register = async (data: TRegister): Promise<TRegisterResponse> => {
  const response = await api.post("/register", data);
  return response.data;
};

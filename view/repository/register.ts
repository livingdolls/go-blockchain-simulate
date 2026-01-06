import { api } from "@/lib/axios";
import { TApiResponse } from "@/types/http";
import { TRegister, TRegisterResponse } from "@/types/register";

export const Register = async (
  data: TRegister
): Promise<TApiResponse<TRegisterResponse>> => {
  const response = await api.post("/register", data);
  return response.data;
};

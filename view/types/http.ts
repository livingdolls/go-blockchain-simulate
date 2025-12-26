export type TApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: string;
};

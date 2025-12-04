import { Wallet } from "ethers";

export const createEthersBackup = async (
  privateKey: string,
  password: string
) => {
  const wallet = new Wallet("0x" + privateKey);

  const json = await wallet.encrypt(password);

  return json;
};

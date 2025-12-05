import { ethers, Wallet } from "ethers";

export const createEthersBackup = async (
  privateKey: string,
  password: string
) => {
  const wallet = new Wallet("0x" + privateKey);

  const json = await wallet.encrypt(password);

  return json;
};

export const walletRestoreFromBackup = async (
  password: string,
  json: string
) => {
  try {
    const wallet = await ethers.Wallet.fromEncryptedJson(json, password);
    return {
      ok: true,
      address: wallet.address,
      wallet: wallet,
    };
  } catch (error) {
    return {
      ok: false,
      error: (error as Error).message,
    };
  }
};

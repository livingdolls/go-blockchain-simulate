import { HDNodeWallet, Wallet } from "ethers";
import { walletRestoreFromBackup } from "./wallet-backup";

export const ValidateWallet = (
  file: File,
  password: string
): Promise<{
  wallet: Wallet | HDNodeWallet | null;
  ok: boolean;
  error?: string;
}> => {
  return new Promise((resolve) => {
    const reader = new FileReader();

    reader.onload = async (e) => {
      try {
        const content = e.target?.result;

        if (typeof content !== "string") {
          resolve({
            wallet: null,
            ok: false,
            error: "File content is not a string",
          });
          return;
        }

        const validate = await walletRestoreFromBackup(password, content);

        if (!validate.ok || !validate.wallet) {
          resolve({
            wallet: null,
            ok: false,
            error: "Invalid wallet or password",
          });
          return;
        }

        resolve({ wallet: validate.wallet, ok: true });
      } catch {
        resolve({
          wallet: null,
          ok: false,
          error: "An error occurred during wallet validation",
        });
      }
    };

    reader.onerror = () => {
      resolve({
        wallet: null,
        ok: false,
        error: "An error occurred while reading the file",
      });
    };

    reader.readAsText(file);
  });
};

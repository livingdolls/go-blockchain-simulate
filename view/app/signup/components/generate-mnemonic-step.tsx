import { Button } from "@/components/ui/button";
import { CardDescription } from "@/components/ui/card";
import { MnemonicWallet } from "@/hooks/use-registration";
import { ArrowLeft, ArrowRight, Download } from "lucide-react";
import { FC } from "react";

type GenerateMnemonicStepProps = {
  onNext: () => void;
  onPrev: () => void;
  generateWallet: () => void;
  wallet: MnemonicWallet;
  handleDownloadBackup: () => void;
};

export const GenerateMnemonicStep: FC<GenerateMnemonicStepProps> = ({
  onNext,
  onPrev,
  generateWallet,
  wallet,
  handleDownloadBackup,
}) => {
  return (
    <div>
      <CardDescription className="mb-4 text-red-600">
        Please back up your mnemonic phrase securely. It is essential for wallet
        recovery.
      </CardDescription>

      {wallet.mnemonic !== "" && (
        <div className="p-4 bg-gray-100 rounded">
          <p className="break-words">{wallet.mnemonic}</p>
        </div>
      )}

      {wallet.mnemonic === "" && (
        <Button type="button" onClick={generateWallet} className="w-full mt-4">
          Generate Mnemonic
        </Button>
      )}

      {wallet.mnemonic !== "" && (
        <div className="flex flex-col gap-2 mt-4">
          <Button
            type="button"
            onClick={handleDownloadBackup}
            className="w-full"
            variant={"outline"}
          >
            Download Wallet Backup
            <Download className="ml-2" />
          </Button>

          <Button type="button" onClick={onNext} className="w-full capitalize">
            I have backed up
            <ArrowRight className="ml-2" />
          </Button>
        </div>
      )}

      <button
        type="button"
        onClick={onPrev}
        className="underline text-sm text-gray-500 mt-2 flex justify-center items-center gap-1 w-full cursor-pointer"
      >
        <ArrowLeft size={12} />
        Prev
      </button>
    </div>
  );
};

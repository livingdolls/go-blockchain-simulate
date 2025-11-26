import { Button } from "@/components/ui/button";
import { CardDescription } from "@/components/ui/card";
import { MnemonicWallet } from "@/hooks/use-registration";
import { FC } from "react";

type GenerateMnemonicStepProps = {
  onNext: () => void;
  onPrev: () => void;
  generateWallet: () => void;
  wallet: MnemonicWallet;
};

export const GenerateMnemonicStep: FC<GenerateMnemonicStepProps> = ({
  onNext,
  onPrev,
  generateWallet,
  wallet,
}) => {
  return (
    <div>
      <CardDescription className="mb-4">
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
        <Button
          type="button"
          onClick={onNext}
          className="w-full mt-2 capitalize"
        >
          I have backed up
        </Button>
      )}

      <button
        type="button"
        onClick={onPrev}
        className="underline text-sm text-gray-500"
      >
        Prev
      </button>
    </div>
  );
};

import { Button } from "@/components/ui/button";
import { FC } from "react";

type FirstStepProps = {
  onNext: () => void;
};

export const FirstStep: FC<FirstStepProps> = ({ onNext }) => {
  return (
    <div>
      <Button type="button" onClick={onNext} className="w-full">
        Next
      </Button>
    </div>
  );
};

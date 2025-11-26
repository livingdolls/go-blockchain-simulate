import { FC } from "react";

type StepContainerProps = {
  step: number;
  currentStep: number;
  children: React.ReactNode;
};

export const StepContainer: FC<StepContainerProps> = ({
  step,
  currentStep,
  children,
}) => {
  return (
    <div className={currentStep === step ? "block" : "hidden"}>{children}</div>
  );
};

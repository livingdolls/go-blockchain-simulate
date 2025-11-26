"use client";

import { StepContainer } from "@/components/step-container";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useRegistration } from "@/hooks/use-registration";
import { FirstStep } from "./components/first-step";
import { SecondStep } from "./components/second-step";
import { GenerateMnemonicStep } from "./components/generate-mnemonic-step";

export default function SignupPage() {
  const {
    currentStep,
    nextStep,
    prevStep,
    generateWallet,
    wallet,
    username,
    onChangeUsername,
    handleSubmitRegistration,
  } = useRegistration();
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <div className="flex flex-col gap-6">
          <Card className="gap-2!">
            <CardHeader className="text-center">
              <CardTitle className="text-xl">Create your Wallet</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmitRegistration}>
                <StepContainer step={1} currentStep={currentStep}>
                  <FirstStep onNext={nextStep} />
                </StepContainer>

                <StepContainer step={2} currentStep={currentStep}>
                  <GenerateMnemonicStep
                    onNext={nextStep}
                    onPrev={prevStep}
                    generateWallet={generateWallet}
                    wallet={wallet}
                  />
                </StepContainer>

                <StepContainer step={3} currentStep={currentStep}>
                  <SecondStep
                    onNext={nextStep}
                    username={username}
                    onChangeUsername={onChangeUsername}
                  />
                </StepContainer>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

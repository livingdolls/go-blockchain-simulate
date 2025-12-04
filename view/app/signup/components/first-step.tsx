import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { FC, useState } from "react";

type FirstStepProps = {
  onNext: () => void;
  onChangeForm: (e: React.ChangeEvent<HTMLInputElement>, field: string) => void;
  password: string;
  repeatPassword: string;
};

export const FirstStep: FC<FirstStepProps> = ({
  onNext,
  onChangeForm,
  password,
  repeatPassword,
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isRepeatVisible, setIsRepeatVisible] = useState(false);
  return (
    <div>
      <Alert variant="destructive" className="mb-4">
        <AlertTitle>
          <span className="font-medium">Important:</span>
        </AlertTitle>

        <AlertDescription>
          Please create a strong password to secure your wallet. Make sure to
          remember it, as it will be required for future access.
        </AlertDescription>
      </Alert>

      <FieldGroup>
        <Field>
          <div className="relative">
            <Input
              type={isVisible ? "text" : "password"}
              placeholder="Password"
              className="pr-9"
              value={password}
              onChange={(e) => onChangeForm(e, "password")}
            />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsVisible((prevState) => !prevState)}
              className="text-muted-foreground focus-visible:ring-ring/50 absolute inset-y-0 right-0 rounded-l-none hover:bg-transparent"
            >
              {isVisible ? <EyeOffIcon /> : <EyeIcon />}
              <span className="sr-only">
                {isVisible ? "Hide password" : "Show password"}
              </span>
            </Button>
          </div>
        </Field>

        <Field>
          <div className="relative">
            <Input
              type={isRepeatVisible ? "text" : "password"}
              placeholder="Repeat Password"
              className="pr-9"
              value={repeatPassword}
              onChange={(e) => onChangeForm(e, "repeatPassword")}
            />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setIsRepeatVisible((prevState) => !prevState)}
              className="text-muted-foreground focus-visible:ring-ring/50 absolute inset-y-0 right-0 rounded-l-none hover:bg-transparent"
            >
              {isRepeatVisible ? <EyeOffIcon /> : <EyeIcon />}
              <span className="sr-only">
                {isRepeatVisible
                  ? "Hide repeat password"
                  : "Show repeat password"}
              </span>
            </Button>
          </div>
        </Field>

        <Button type="button" onClick={onNext} className="w-full">
          Next
        </Button>
      </FieldGroup>
    </div>
  );
};

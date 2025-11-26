import { Button } from "@/components/ui/button";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { FC } from "react";

type SecondStepProps = {
  onNext: () => void;
  username: string;
  onChangeUsername: (e: React.ChangeEvent<HTMLInputElement>) => void;
};

export const SecondStep: FC<SecondStepProps> = ({
  onNext,
  username,
  onChangeUsername,
}) => {
  return (
    <div>
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="username">Username</FieldLabel>
          <Input
            id="username"
            type="text"
            placeholder="John Doe"
            required
            value={username}
            onChange={onChangeUsername}
          />
        </Field>

        <Button className="w-full">Create Wallet</Button>
      </FieldGroup>
    </div>
  );
};

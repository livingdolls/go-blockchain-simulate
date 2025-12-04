import { Button } from "@/components/ui/button";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { FC } from "react";

type SecondStepProps = {
  username: string;
  onChangeUsername: (
    e: React.ChangeEvent<HTMLInputElement>,
    field: string
  ) => void;
};

export const SecondStep: FC<SecondStepProps> = ({
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
            onChange={(e) => onChangeUsername(e, "username")}
          />
        </Field>

        <Button className="w-full">Create Wallet</Button>
      </FieldGroup>
    </div>
  );
};

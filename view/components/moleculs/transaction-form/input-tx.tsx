import { Input } from "@/components/ui/input";

type Props = {
  label: string;
  name: string;
  type: string;
  placeholder?: string;
  value: string | number;
  onChange: (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => void;
  disabled?: boolean;
};

export const InputTx: React.FC<Props> = ({
  label,
  name,
  type,
  placeholder,
  value,
  onChange,
  disabled,
}) => {
  return (
    <div className="flex flex-col">
      <label htmlFor={name} className="mb-1 font-medium">
        {label}
      </label>
      <Input
        type={type}
        id={name}
        name={name}
        className="rounded border border-gray-300 p-2"
        placeholder={placeholder}
        onChange={onChange}
        value={value}
        disabled={disabled}
      />
    </div>
  );
};

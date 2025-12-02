import { Textarea } from "@/components/ui/textarea";

type Props = {
  label: string;
  name: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  disabled?: boolean;
};

export const TextareaTx: React.FC<Props> = ({
  label,
  name,
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
      <Textarea
        id={name}
        name={name}
        className="rounded border border-gray-300 p-2"
        placeholder={placeholder}
        onChange={onChange}
        value={value}
        // spatial media crunch crop clump candy rotate hollow amount tissue total scene
        //fancy pair mammal swarm they syrup discover school rug obtain extend hotel
        rows={3}
        disabled={disabled}
      />
    </div>
  );
};

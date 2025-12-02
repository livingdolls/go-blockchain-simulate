import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type Props = {
  data: string[];
  setFilterValue: (name: string, value: string) => void;
  placeholder: string;
  label: string;
  value: string;
};

export const SelectFilter = ({
  data,
  setFilterValue,
  placeholder,
  label,
  value,
}: Props) => {
  return (
    <>
      <Select
        onValueChange={(value) => setFilterValue("status", value)}
        value={value}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>{label}</SelectLabel>
            {data.map((status) => (
              <SelectItem key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1).toLowerCase()}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </>
  );
};

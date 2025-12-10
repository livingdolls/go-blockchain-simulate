import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";

type WalletFileDropzoneProps = {
  onFile: (file: File, content: unknown) => void;
  disabled?: boolean;
};

export const WalletFileDropzone: React.FC<WalletFileDropzoneProps> = ({
  onFile,
  disabled,
}) => {
  const [fileName, setFileName] = useState<string | null>(null);

  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        const file = acceptedFiles[0];
        setFileName(file.name);

        const reader = new FileReader();
        reader.onload = (e) => {
          try {
            const text = e.target?.result as string;
            const json = JSON.parse(text);
            onFile(file, json);
          } catch (error) {
            console.error("Error reading wallet file:", error);
            onFile(file, null);
          }
        };
        reader.readAsText(file);
      }
    },
    [onFile]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { "application/json": [".json"] },
    multiple: false,
    disabled,
  });

  return (
    <div
      {...getRootProps()}
      className={`border border-dashed rounded p-4 text-center cursor-pointer ${
        isDragActive ? "bg-gray-100" : ""
      } ${disabled ? "opacity-50 pointer-events-none" : ""}`}
    >
      <input {...getInputProps()} />
      {fileName ? (
        <p className="text-sm text-gray-700">{fileName}</p>
      ) : (
        <p className="text-sm text-gray-500">
          {isDragActive
            ? "Drop the wallet file here..."
            : "Drag & drop wallet file (JSON) here, or click to select"}
        </p>
      )}
    </div>
  );
};

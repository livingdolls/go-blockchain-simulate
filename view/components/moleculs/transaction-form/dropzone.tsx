import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";

type WalletFileDropzoneProps = {
  onFile: (file: File) => void;
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
        onFile(acceptedFiles[0]);
        setFileName(acceptedFiles[0].name);
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

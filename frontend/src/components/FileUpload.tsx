import { useState, useCallback } from 'react';
import { Upload, File, X, Loader2 } from 'lucide-react';

interface FileUploadProps {
  onUpload: (file: File) => void;
  isLoading: boolean;
}

export function FileUpload({ onUpload, isLoading }: FileUploadProps) {
  const [isDragOver, setIsDragOver] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);

    const file = e.dataTransfer.files[0];
    if (file && file.name.endsWith('.csv')) {
      setSelectedFile(file);
    }
  }, []);

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  }, []);

  const handleSubmit = useCallback(() => {
    if (selectedFile) {
      onUpload(selectedFile);
      setSelectedFile(null);
    }
  }, [selectedFile, onUpload]);

  const clearFile = useCallback(() => {
    setSelectedFile(null);
  }, []);

  return (
    <div className="w-full">
      {!selectedFile ? (
        <div
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          className={`
            relative border-2 border-dashed rounded-2xl p-12 text-center
            transition-all duration-300 cursor-pointer
            ${isDragOver 
              ? 'border-indigo-500 bg-indigo-50/50 dark:bg-indigo-950/30' 
              : 'border-[var(--glass-border)] hover:border-indigo-400'
            }
          `}
        >
          <input
            type="file"
            accept=".csv"
            onChange={handleFileSelect}
            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          />
          
          <div className="flex flex-col items-center gap-4">
            <div className={`
              w-20 h-20 rounded-2xl flex items-center justify-center
              transition-all duration-300
              ${isDragOver 
                ? 'bg-indigo-500 text-white scale-110' 
                : 'bg-[var(--glass-bg)] text-[var(--text-secondary)]'
              }
            `}>
              {isLoading ? (
                <Loader2 className="w-8 h-8 animate-spin" />
              ) : (
                <Upload className="w-8 h-8" />
              )}
            </div>
            
            <div>
              <p className="text-lg font-medium text-[var(--text-primary)]">
                {isDragOver ? 'Drop your CSV here' : 'Drag & drop your CSV file'}
              </p>
              <p className="text-sm text-[var(--text-secondary)] mt-1">
                or click to browse files
              </p>
            </div>
            
            <p className="text-xs text-[var(--text-secondary)]">
              Supports .csv files only
            </p>
          </div>
        </div>
      ) : (
        <div className="glass-card p-6">
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
              <File className="w-7 h-7 text-white" />
            </div>
            
            <div className="flex-1 min-w-0">
              <p className="font-medium text-[var(--text-primary)] truncate">
                {selectedFile.name}
              </p>
              <p className="text-sm text-[var(--text-secondary)]">
                {(selectedFile.size / 1024).toFixed(1)} KB
              </p>
            </div>
            
            <button
              onClick={clearFile}
              disabled={isLoading}
              className="p-2 rounded-lg hover:bg-red-500/10 text-[var(--text-secondary)] hover:text-red-500 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
          
          <button
            onClick={handleSubmit}
            disabled={isLoading}
            className="btn-primary w-full mt-4 flex items-center justify-center gap-2"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Uploading...
              </>
            ) : (
              <>
                <Upload className="w-5 h-5" />
                Upload File
              </>
            )}
          </button>
        </div>
      )}
    </div>
  );
}

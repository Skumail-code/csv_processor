export interface NullString {
  String: string;
  Valid: boolean;
}

export interface Job {
  jobId: string;
  status: 'queued' | 'processing' | 'done' | 'failed';
  totalRows: number;
  processedRows: number;
  invalidRows: number;
  errorMessage?: NullString;
  createdAt: string;
  updatedAt: string;
}

export interface UploadResponse {
  jobId: string;
  status: string;
  task_id: string;
}

export interface JobStatusResponse {
  jobId: string;
  status: string;
  totalRows: number;
  processedRows: number;
  invalidRows: number;
  errorMessage?: NullString;
  createdAt: string;
  updatedAt: string;
}

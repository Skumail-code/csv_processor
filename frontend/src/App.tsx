import { useState, useEffect, useCallback } from 'react';
import { Upload, FileText, Moon, Sun, Activity, CheckCircle, AlertCircle, Clock } from 'lucide-react';
import axios from 'axios';
import { useTheme } from './hooks/useTheme';
import { Job, UploadResponse, JobStatusResponse } from './types';
import { FileUpload } from './components/FileUpload';
import { JobCard } from './components/JobCard';
import './App.css';

const api = axios.create({
  baseURL: '/api',
});

function App() {
  const { isDark, toggle } = useTheme();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Poll for job status updates
  useEffect(() => {
    const pollJobs = async () => {
      const activeJobs = jobs.filter(j => j.status === 'queued' || j.status === 'processing');
      
      for (const job of activeJobs) {
        try {
          const { data } = await api.get<JobStatusResponse>(`/status/${job.jobId}`);
          
          setJobs(prev => prev.map(j => 
            j.jobId === data.jobId 
              ? { ...j, ...data, status: data.status as Job['status'] }
              : j
          ));
        } catch (err) {
          console.error('Failed to poll job status:', err);
        }
      }
    };

    const interval = setInterval(pollJobs, 2000);
    return () => clearInterval(interval);
  }, [jobs]);

  const handleUpload = useCallback(async (file: File) => {
    setIsLoading(true);
    setError(null);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const { data } = await api.post<UploadResponse>('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });

      const newJob: Job = {
        jobId: data.jobId,
        status: 'queued',
        totalRows: 0,
        processedRows: 0,
        invalidRows: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      setJobs(prev => [newJob, ...prev]);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to upload file');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleDownload = useCallback(async (jobId: string, type: 'processed' | 'damaged') => {
    try {
      const endpoint = type === 'processed' 
        ? `/download/${jobId}` 
        : `/download/${jobId}/damaged`;
      
      const response = await api.get(endpoint, { responseType: 'blob' });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', type === 'processed' ? 'processed_transactions.csv' : 'damaged_rows.csv');
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to download ${type} file`);
    }
  }, []);

  const removeJob = useCallback((jobId: string) => {
    setJobs(prev => prev.filter(j => j.jobId !== jobId));
  }, []);

  const getStatusIcon = (status: Job['status']) => {
    switch (status) {
      case 'done':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'processing':
        return <Activity className="w-5 h-5 text-amber-500 animate-pulse" />;
      case 'failed':
        return <AlertCircle className="w-5 h-5 text-red-500" />;
      default:
        return <Clock className="w-5 h-5 text-slate-500" />;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[var(--bg-primary)] to-[var(--bg-secondary)]">
      {/* Header */}
      <header className="glass sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg">
              <FileText className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold gradient-text">CSV Processor</h1>
              <p className="text-xs text-[var(--text-secondary)]">Process & analyze transaction files</p>
            </div>
          </div>
          
          <button
            onClick={toggle}
            className="p-2 rounded-xl glass hover:bg-[var(--glass-bg)] transition-colors"
            aria-label="Toggle theme"
          >
            {isDark ? (
              <Sun className="w-5 h-5 text-amber-400" />
            ) : (
              <Moon className="w-5 h-5 text-slate-600" />
            )}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* Upload Section */}
        <section className="mb-12">
          <div className="glass-card p-8">
            <div className="text-center mb-6">
              <h2 className="text-2xl font-semibold mb-2 text-[var(--text-primary)]">
                Upload CSV File
              </h2>
              <p className="text-[var(--text-secondary)]">
                Drag and drop or click to select your transaction CSV file
              </p>
            </div>
            
            <FileUpload onUpload={handleUpload} isLoading={isLoading} />
            
            {error && (
              <div className="mt-4 p-4 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-3">
                <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0" />
                <p className="text-red-600 dark:text-red-400 text-sm">{error}</p>
              </div>
            )}
          </div>
        </section>

        {/* Jobs List */}
        <section>
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-[var(--text-primary)]">
              Processing Jobs
            </h2>
            <span className="text-sm text-[var(--text-secondary)]">
              {jobs.length} job{jobs.length !== 1 ? 's' : ''}
            </span>
          </div>

          {jobs.length === 0 ? (
            <div className="glass-card p-12 text-center">
              <div className="w-16 h-16 rounded-full bg-[var(--glass-bg)] flex items-center justify-center mx-auto mb-4">
                <Upload className="w-8 h-8 text-[var(--text-secondary)]" />
              </div>
              <p className="text-[var(--text-secondary)]">
                No jobs yet. Upload a CSV file to get started.
              </p>
            </div>
          ) : (
            <div className="grid gap-4">
              {jobs.map((job) => (
                <JobCard
                  key={job.jobId}
                  job={job}
                  onDownload={handleDownload}
                  onRemove={removeJob}
                  getStatusIcon={getStatusIcon}
                />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;

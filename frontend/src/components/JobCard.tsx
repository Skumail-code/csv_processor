import { Download, Trash2, FileX, Calendar, BarChart3 } from 'lucide-react';
import { format } from 'date-fns';
import { Job } from '../types';
import { ReactNode } from 'react';

interface JobCardProps {
  job: Job;
  onDownload: (jobId: string, type: 'processed' | 'damaged') => void;
  onRemove: (jobId: string) => void;
  getStatusIcon: (status: Job['status']) => ReactNode;
}

export function JobCard({ job, onDownload, onRemove, getStatusIcon }: JobCardProps) {
  const progress = job.totalRows > 0 
    ? Math.round((job.processedRows / job.totalRows) * 100) 
    : 0;

  const getStatusClass = (status: Job['status']) => {
    switch (status) {
      case 'done': return 'status-done';
      case 'processing': return 'status-processing';
      case 'failed': return 'status-failed';
      default: return 'status-queued';
    }
  };

  return (
    <div className="glass-card p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-4 flex-1 min-w-0">
          <div className="flex-shrink-0">
            {getStatusIcon(job.status)}
          </div>
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1">
              <h3 className="font-semibold text-[var(--text-primary)] truncate">
                Job {job.jobId.slice(0, 8)}...
              </h3>
              <span className={getStatusClass(job.status)}>
                {job.status}
              </span>
            </div>
            
            <div className="flex items-center gap-4 text-sm text-[var(--text-secondary)] mb-3">
              <span className="flex items-center gap-1">
                <Calendar className="w-4 h-4" />
                {format(new Date(job.createdAt), 'MMM d, h:mm a')}
              </span>
              {job.totalRows > 0 && (
                <span className="flex items-center gap-1">
                  <BarChart3 className="w-4 h-4" />
                  {job.processedRows.toLocaleString()} / {job.totalRows.toLocaleString()} rows
                </span>
              )}
            </div>

            {/* Progress Bar - shows for both queued and processing */}
            {(job.status === 'queued' || job.status === 'processing') && (
              <div className="mb-3">
                <div className="flex justify-between text-xs mb-1">
                  <span className="text-[var(--text-secondary)] flex items-center gap-1">
                    {job.status === 'queued' ? (
                      <>
                        <span className="w-2 h-2 bg-amber-400 rounded-full animate-pulse" />
                        Waiting in queue...
                      </>
                    ) : (
                      <>
                        <span className="w-2 h-2 bg-indigo-500 rounded-full animate-pulse" />
                        Processing rows...
                      </>
                    )}
                  </span>
                  {job.status === 'processing' && job.totalRows > 0 && (
                    <span className="text-[var(--accent)] font-medium">{progress}%</span>
                  )}
                </div>
                <div className="h-2 bg-[var(--glass-bg)] rounded-full overflow-hidden">
                  <div 
                    className={`
                      h-full rounded-full transition-all duration-500
                      ${job.status === 'queued' 
                        ? 'bg-gradient-to-r from-amber-400 to-orange-400 w-1/3 animate-pulse' 
                        : 'bg-gradient-to-r from-indigo-500 to-purple-500'}
                    `}
                    style={{ width: job.status === 'processing' ? `${progress}%` : undefined }}
                  />
                </div>
                {job.status === 'processing' && job.totalRows > 0 && (
                  <p className="text-xs text-[var(--text-secondary)] mt-1">
                    {job.processedRows.toLocaleString()} of {job.totalRows.toLocaleString()} rows processed
                  </p>
                )}
              </div>
            )}

            {/* Stats */}
            {job.status === 'done' && (
              <div className="flex items-center gap-4 text-sm">
                <span className="text-emerald-600 dark:text-emerald-400 font-medium">
                  {job.totalRows - job.invalidRows} valid rows
                </span>
                {job.invalidRows > 0 && (
                  <span className="text-amber-600 dark:text-amber-400 font-medium">
                    {job.invalidRows} damaged rows
                  </span>
                )}
              </div>
            )}

            {job.errorMessage?.Valid && job.errorMessage.String && (
              <p className="text-sm text-red-500 mt-2 line-clamp-2">
                {job.errorMessage.String}
              </p>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 flex-shrink-0">
          {job.status === 'done' && (
            <>
              <button
                onClick={() => onDownload(job.jobId, 'processed')}
                className="btn-secondary flex items-center gap-2 text-sm py-2 px-3"
                title="Download processed file"
              >
                <Download className="w-4 h-4" />
                <span className="hidden sm:inline">Result</span>
              </button>
              
              {job.invalidRows > 0 && (
                <button
                  onClick={() => onDownload(job.jobId, 'damaged')}
                  className="btn-secondary flex items-center gap-2 text-sm py-2 px-3 text-amber-600 dark:text-amber-400"
                  title="Download damaged rows"
                >
                  <FileX className="w-4 h-4" />
                  <span className="hidden sm:inline">Damaged</span>
                </button>
              )}
            </>
          )}
          
          <button
            onClick={() => onRemove(job.jobId)}
            className="p-2 rounded-lg hover:bg-red-500/10 text-[var(--text-secondary)] hover:text-red-500 transition-colors"
            title="Remove from list"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}

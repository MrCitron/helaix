import React from 'react';
import { useI18n } from '../i18n';
import { GxOpenPath, GxOpenFolderOfFile } from '../../wailsjs/go/main/App';

const ExportModal = ({ isOpen, onClose, filePath }) => {
    const { t } = useI18n();

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
            <div className="w-full max-w-lg bg-white dark:bg-surface-dark border border-slate-300 dark:border-[#3f5256] rounded-2xl shadow-2xl relative overflow-hidden animate-in zoom-in duration-300">
                <div className="p-8 flex flex-col items-center">
                    <div className="w-16 h-16 rounded-full bg-[#13c8ec]/10 flex items-center justify-center mb-6 shadow-[0_0_20px_rgba(19,200,236,0.1)]">
                        <span className="material-symbols-outlined text-[#13c8ec] text-4xl">check</span>
                    </div>

                    <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-2 font-display text-center">
                        {t('exportModal.title')}
                    </h3>

                    <p className="text-slate-600 dark:text-text-secondary text-center text-sm mb-8 font-body leading-relaxed">
                        {t('exportModal.description')}
                    </p>

                    <div className="w-full bg-slate-100 dark:bg-background-dark border border-slate-300 dark:border-border-dark rounded-xl p-4 mb-8">
                        <div className="flex items-center gap-3 font-mono text-xs text-slate-700 dark:text-gray-300 bg-white dark:bg-[#0b0f10] p-3 rounded-lg border border-slate-200 dark:border-surface-dark overflow-hidden">
                            <span className="material-symbols-outlined text-slate-500 dark:text-text-muted text-lg shrink-0">description</span>
                            <span className="truncate" title={filePath}>{filePath}</span>
                        </div>
                    </div>

                    <button
                        onClick={onClose}
                        className="w-full bg-[#13c8ec] hover:bg-[#0fb3d4] text-slate-900 font-bold py-3.5 rounded-xl transition-all shadow-[0_0_15px_rgba(19,200,236,0.25)] hover:shadow-[0_0_25px_rgba(19,200,236,0.4)] font-display text-sm tracking-wide"
                    >
                        {t('exportModal.close')}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default ExportModal;
